// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package group

import (
	"bytes"
	"fmt"
	"unsafe"

	batch "github.com/matrixorigin/matrixone/pkg/container/batch2"
	"github.com/matrixorigin/matrixone/pkg/container/hashtable"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/ring"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec2/aggregate"
	"github.com/matrixorigin/matrixone/pkg/vectorize/add"
	process "github.com/matrixorigin/matrixone/pkg/vm/process2"
)

func String(arg interface{}, buf *bytes.Buffer) {
	ap := arg.(*Argument)
	buf.WriteString("γ([")
	for i, pos := range ap.Poses {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v", pos))
	}
	buf.WriteString("], [")
	for i, agg := range ap.Aggs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v(%v)", aggregate.Names[agg.Op], agg.Pos))
	}
	buf.WriteString("])")
}

func Prepare(_ *process.Process, arg interface{}) error {
	ap := arg.(*Argument)
	ap.ctr = new(Container)
	return nil
}

func Call(proc *process.Process, arg interface{}) (bool, error) {
	ap := arg.(*Argument)
	if len(ap.Poses) == 0 {
		return ap.ctr.process(ap, proc)
	}
	return ap.ctr.processWithGroup(ap, proc)
}

func (ctr *Container) process(ap *Argument, proc *process.Process) (bool, error) {
	bat := proc.Reg.InputBatch
	if bat == nil {
		if ctr.bat != nil {
			proc.Reg.InputBatch = ctr.bat
			ctr.bat = nil
		}
		return true, nil
	}
	if len(bat.Zs) == 0 {
		return false, nil
	}
	defer batch.Clean(bat, proc.Mp)
	proc.Reg.InputBatch = &batch.Batch{}
	if ctr.bat == nil {
		var err error

		ctr.bat = batch.New(0)
		ctr.bat.Zs = []int64{0}
		ctr.bat.Rs = make([]ring.Ring, len(ap.Aggs))
		for i, agg := range ap.Aggs {
			if ctr.bat.Rs[i], err = aggregate.New(agg.Op, bat.Vecs[agg.Pos].Typ); err != nil {
				return false, err
			}
		}
		for _, r := range ctr.bat.Rs {
			if err := r.Grow(proc.Mp); err != nil {
				batch.Clean(ctr.bat, proc.Mp)
				return false, err
			}
		}
	}
	if err := ctr.processH0(bat, ap, proc); err != nil {
		batch.Clean(ctr.bat, proc.Mp)
		return false, err
	}
	return false, nil
}

func (ctr *Container) processWithGroup(ap *Argument, proc *process.Process) (bool, error) {
	var err error

	bat := proc.Reg.InputBatch
	if bat == nil {
		if ctr.bat != nil {
			switch ctr.typ {
			case H8:
				ctr.bat.Ht = ctr.intHashMap
			case H24:
				ctr.bat.Ht = ctr.strHashMap
			case H32:
				ctr.bat.Ht = ctr.strHashMap
			case H40:
				ctr.bat.Ht = ctr.strHashMap
			default:
				ctr.bat.Ht = ctr.strHashMap
			}
			proc.Reg.InputBatch = ctr.bat
			ctr.bat = nil
		}
		return true, nil
	}
	if len(bat.Zs) == 0 {
		return false, nil
	}
	defer batch.Clean(bat, proc.Mp)
	proc.Reg.InputBatch = &batch.Batch{}
	if ctr.bat == nil {
		size := 0
		ctr.bat = batch.New(len(ap.Poses))
		for i, pos := range ap.Poses {
			vec := bat.Vecs[pos]
			ctr.bat.Vecs[i] = vector.New(vec.Typ)
			switch vec.Typ.Oid {
			case types.T_int8, types.T_uint8:
				size += 1 + 1
			case types.T_int16, types.T_uint16:
				size += 2 + 1
			case types.T_int32, types.T_uint32, types.T_float32, types.T_date:
				size += 4 + 1
			case types.T_int64, types.T_uint64, types.T_float64, types.T_datetime, types.T_decimal64:
				size += 8 + 1
			case types.T_decimal128:
				size += 16 + 1
			case types.T_char, types.T_varchar:
				if width := vec.Typ.Width; width > 0 {
					size += int(width) + 1
				} else {
					size = 128
				}
			}
		}
		ctr.bat.Rs = make([]ring.Ring, len(ap.Aggs))
		for i, agg := range ap.Aggs {
			if ctr.bat.Rs[i], err = aggregate.New(agg.Op, bat.Vecs[agg.Pos].Typ); err != nil {
				return false, err
			}
		}
		ctr.keyOffs = make([]uint32, UnitLimit)
		ctr.zKeyOffs = make([]uint32, UnitLimit)
		ctr.inserted = make([]uint8, UnitLimit)
		ctr.zInserted = make([]uint8, UnitLimit)
		ctr.hashes = make([]uint64, UnitLimit)
		ctr.strHashStates = make([][3]uint64, UnitLimit)
		ctr.values = make([]uint64, UnitLimit)
		ctr.intHashMap = &hashtable.Int64HashMap{}
		ctr.strHashMap = &hashtable.StringHashMap{}
		switch {
		case size <= 8:
			ctr.typ = H8
			ctr.h8.keys = make([]uint64, UnitLimit)
			ctr.h8.zKeys = make([]uint64, UnitLimit)
			ctr.intHashMap.Init()
		case size <= 24:
			ctr.typ = H24
			ctr.h24.keys = make([][3]uint64, UnitLimit)
			ctr.h24.zKeys = make([][3]uint64, UnitLimit)
			ctr.strHashMap.Init()
		case size <= 32:
			ctr.typ = H32
			ctr.h32.keys = make([][4]uint64, UnitLimit)
			ctr.h32.zKeys = make([][4]uint64, UnitLimit)
			ctr.strHashMap.Init()
		case size <= 40:
			ctr.typ = H40
			ctr.h40.keys = make([][5]uint64, UnitLimit)
			ctr.h40.zKeys = make([][5]uint64, UnitLimit)
			ctr.strHashMap.Init()
		default:
			ctr.typ = HStr
			ctr.hstr.keys = make([][]byte, UnitLimit)
			ctr.strHashMap.Init()
		}
	}
	switch ctr.typ {
	case H8:
		err = ctr.processH8(bat, ap, proc)
	case H24:
		err = ctr.processH24(bat, ap, proc)
	case H32:
		err = ctr.processH32(bat, ap, proc)
	case H40:
		err = ctr.processH40(bat, ap, proc)
	default:
		err = ctr.processHStr(bat, ap, proc)
	}
	if err != nil {
		batch.Clean(ctr.bat, proc.Mp)
		ctr.bat = nil
		return false, err
	}
	return false, err
}

func (ctr *Container) processH0(bat *batch.Batch, ap *Argument, proc *process.Process) error {
	for _, z := range bat.Zs {
		ctr.bat.Zs[0] += z
	}
	for i, r := range ctr.bat.Rs {
		r.BulkFill(0, bat.Zs, bat.Vecs[ap.Aggs[i].Pos])
	}
	return nil
}

func (ctr *Container) processH8(bat *batch.Batch, ap *Argument, proc *process.Process) error {
	count := len(bat.Zs)
	for i := 0; i < count; i += UnitLimit {
		n := count - i
		if n > UnitLimit {
			n = UnitLimit
		}
		copy(ctr.keyOffs, ctr.zKeyOffs)
		copy(ctr.h8.keys, ctr.h8.zKeys)
		for _, pos := range ap.Poses {
			vec := bat.Vecs[pos]
			switch typLen := vec.Typ.Oid.FixedLength(); typLen {
			case 1:
				fillGroup[uint8](ctr, vec, ctr.h8.keys, n, 1, i)
			case 2:
				fillGroup[uint16](ctr, vec, ctr.h8.keys, n, 2, i)
			case 4:
				fillGroup[uint32](ctr, vec, ctr.h8.keys, n, 4, i)
			case 8:
				fillGroup[uint64](ctr, vec, ctr.h8.keys, n, 8, i)
			case -8:
				fillGroup[uint64](ctr, vec, ctr.h8.keys, n, 8, i)
			case -16:
				fillGroup[types.Decimal128](ctr, vec, ctr.h8.keys, n, 16, i)
			default:
				fillStringGroup(ctr, vec, ctr.h8.keys, n, 8, i)
			}
		}
		ctr.hashes[0] = 0
		ctr.intHashMap.InsertBatch(n, ctr.hashes, unsafe.Pointer(&ctr.h8.keys[0]), ctr.values)
		if err := ctr.batchFill(i, n, bat, ap, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) processH24(bat *batch.Batch, ap *Argument, proc *process.Process) error {
	count := len(bat.Zs)
	for i := 0; i < count; i += UnitLimit {
		n := count - i
		if n > UnitLimit {
			n = UnitLimit
		}
		copy(ctr.keyOffs, ctr.zKeyOffs)
		copy(ctr.h24.keys, ctr.h24.zKeys)
		for _, pos := range ap.Poses {
			vec := bat.Vecs[pos]
			switch typLen := vec.Typ.Oid.FixedLength(); typLen {
			case 1:
				fillGroup[uint8](ctr, vec, ctr.h24.keys, n, 1, i)
			case 2:
				fillGroup[uint16](ctr, vec, ctr.h24.keys, n, 2, i)
			case 4:
				fillGroup[uint32](ctr, vec, ctr.h24.keys, n, 4, i)
			case 8:
				fillGroup[uint64](ctr, vec, ctr.h24.keys, n, 8, i)
			case -8:
				fillGroup[types.Decimal64](ctr, vec, ctr.h24.keys, n, 8, i)
			case -16:
				fillGroup[types.Decimal128](ctr, vec, ctr.h24.keys, n, 16, i)
			default:
				fillStringGroup(ctr, vec, ctr.h24.keys, n, 24, i)
			}
		}
		ctr.strHashMap.InsertString24Batch(ctr.strHashStates, ctr.h24.keys[:n], ctr.values)
		if err := ctr.batchFill(i, n, bat, ap, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) processH32(bat *batch.Batch, ap *Argument, proc *process.Process) error {
	count := len(bat.Zs)
	for i := 0; i < count; i += UnitLimit {
		n := count - i
		if n > UnitLimit {
			n = UnitLimit
		}
		copy(ctr.keyOffs, ctr.zKeyOffs)
		copy(ctr.h32.keys, ctr.h32.zKeys)
		for _, pos := range ap.Poses {
			vec := bat.Vecs[pos]
			switch typLen := vec.Typ.Oid.FixedLength(); typLen {
			case 1:
				fillGroup[uint8](ctr, vec, ctr.h32.keys, n, 1, i)
			case 2:
				fillGroup[uint16](ctr, vec, ctr.h32.keys, n, 2, i)
			case 4:
				fillGroup[uint32](ctr, vec, ctr.h32.keys, n, 4, i)
			case 8:
				fillGroup[uint64](ctr, vec, ctr.h32.keys, n, 8, i)
			case -8:
				fillGroup[uint64](ctr, vec, ctr.h32.keys, n, 8, i)
			case -16:
				fillGroup[types.Decimal128](ctr, vec, ctr.h32.keys, n, 16, i)
			default:
				fillStringGroup(ctr, vec, ctr.h32.keys, n, 32, i)
			}
		}
		ctr.strHashMap.InsertString32Batch(ctr.strHashStates, ctr.h32.keys[:n], ctr.values)
		if err := ctr.batchFill(i, n, bat, ap, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) processH40(bat *batch.Batch, ap *Argument, proc *process.Process) error {
	count := len(bat.Zs)
	for i := 0; i < count; i += UnitLimit {
		n := count - i
		if n > UnitLimit {
			n = UnitLimit
		}
		copy(ctr.keyOffs, ctr.zKeyOffs)
		copy(ctr.h40.keys, ctr.h40.zKeys)
		for _, pos := range ap.Poses {
			vec := bat.Vecs[pos]
			switch typLen := vec.Typ.Oid.FixedLength(); typLen {
			case 1:
				fillGroup[uint8](ctr, vec, ctr.h40.keys, n, 1, i)
			case 2:
				fillGroup[uint16](ctr, vec, ctr.h40.keys, n, 2, i)
			case 4:
				fillGroup[uint32](ctr, vec, ctr.h40.keys, n, 4, i)
			case 8:
				fillGroup[uint64](ctr, vec, ctr.h40.keys, n, 8, i)
			case -8:
				fillGroup[uint64](ctr, vec, ctr.h40.keys, n, 8, i)
			case -16:
				fillGroup[types.Decimal128](ctr, vec, ctr.h40.keys, n, 16, i)
			default:
				fillStringGroup(ctr, vec, ctr.h40.keys, n, 40, i)
			}
		}
		ctr.strHashMap.InsertString40Batch(ctr.strHashStates, ctr.h40.keys[:n], ctr.values)
		if err := ctr.batchFill(i, n, bat, ap, proc); err != nil {
			return err
		}
	}
	return nil
}

func (ctr *Container) processHStr(bat *batch.Batch, ap *Argument, proc *process.Process) error {
	count := len(bat.Zs)
	for i := 0; i < count; i += UnitLimit { // batch
		n := count - i
		if n > UnitLimit {
			n = UnitLimit
		}
		for _, pos := range ap.Poses {
			vec := bat.Vecs[pos]
			switch typLen := vec.Typ.Oid.FixedLength(); typLen {
			case 1:
				fillGroupStr[uint8](ctr, vec, n, 1, i)
			case 2:
				fillGroupStr[uint16](ctr, vec, n, 2, i)
			case 4:
				fillGroupStr[uint32](ctr, vec, n, 4, i)
			case 8:
				fillGroupStr[uint64](ctr, vec, n, 8, i)
			case -8:
				fillGroupStr[uint64](ctr, vec, n, 8, i)
			case -16:
				fillGroupStr[types.Decimal128](ctr, vec, n, 16, i)
			default:
				vs := vec.Col.(*types.Bytes)
				if !nulls.Any(vec.Nsp) {
					for k := 0; k < n; k++ {
						ctr.hstr.keys[k] = append(ctr.hstr.keys[k], byte(0))
						ctr.hstr.keys[k] = append(ctr.hstr.keys[k], vs.Get(int64(i+k))...)
					}
				} else {
					for k := 0; k < n; k++ {
						if vec.Nsp.Np.Contains(uint64(i + k)) {
							ctr.hstr.keys[k] = append(ctr.hstr.keys[k], byte(1))
						} else {
							ctr.hstr.keys[k] = append(ctr.hstr.keys[k], byte(0))
							ctr.hstr.keys[k] = append(ctr.hstr.keys[k], vs.Get(int64(i+k))...)
						}
					}
				}

			}
		}
		for k := 0; k < n; k++ {
			if l := len(ctr.hstr.keys[k]); l < 16 {
				ctr.hstr.keys[k] = append(ctr.hstr.keys[k], hashtable.StrKeyPadding[l:]...)
			}
		}
		ctr.strHashMap.InsertStringBatch(ctr.strHashStates, ctr.hstr.keys[:n], ctr.values)
		if err := ctr.batchFill(i, n, bat, ap, proc); err != nil {
			return err
		}
		for k := 0; k < n; k++ {
			ctr.hstr.keys[k] = ctr.hstr.keys[k][:0]
		}
	}
	return nil
}

func (ctr *Container) batchFill(i int, n int, bat *batch.Batch, ap *Argument, proc *process.Process) error {
	cnt := 0
	copy(ctr.inserted[:n], ctr.zInserted[:n])
	for k, v := range ctr.values[:n] {
		if v > ctr.rows {
			ctr.inserted[k] = 1
			ctr.rows++
			cnt++
			ctr.bat.Zs = append(ctr.bat.Zs, 0)
		}
		ai := int64(v) - 1
		ctr.bat.Zs[ai] += bat.Zs[i+k]
	}
	if cnt > 0 {
		for j, vec := range ctr.bat.Vecs {
			if err := vector.UnionBatch(vec, bat.Vecs[ap.Poses[j]], int64(i), cnt, ctr.inserted[:n], proc.Mp); err != nil {
				return err
			}
		}
		for _, r := range ctr.bat.Rs {
			if err := r.Grows(cnt, proc.Mp); err != nil {
				return err
			}
		}
	}
	for j, r := range ctr.bat.Rs {
		r.BatchFill(int64(i), ctr.inserted[:n], ctr.values, bat.Zs, bat.Vecs[ap.Aggs[j].Pos])
	}
	return nil
}

func fillGroup[T1, T2 any](ctr *Container, vec *vector.Vector, keys []T2, n int, sz uint32, start int) {
	vs := vector.DecodeFixedCol[T1](vec, int(sz))
	if !nulls.Any(vec.Nsp) {
		for i := 0; i < n; i++ {
			*(*int8)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i])) = 0
			*(*T1)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i]+1)) = vs[i+start]
		}
		add.Uint32AddScalar(1+sz, ctr.keyOffs[:n], ctr.keyOffs[:n])
	} else {
		for i := 0; i < n; i++ {
			if vec.Nsp.Np.Contains(uint64(i + start)) {
				*(*int8)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i])) = 1
				ctr.keyOffs[i]++
			} else {
				*(*int8)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i])) = 0
				*(*T1)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i]+1)) = vs[i+start]
				ctr.keyOffs[i] += 1 + sz
			}
		}
	}
}

func fillStringGroup[T any](ctr *Container, vec *vector.Vector, keys []T, n int, sz uint32, start int) {
	vs := vec.Col.(*types.Bytes)
	vData := vs.Data
	vOff := vs.Offsets
	vLen := vs.Lengths
	if !nulls.Any(vec.Nsp) {
		for i := 0; i < n; i++ {
			*(*int8)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i])) = 0
			copy(unsafe.Slice((*byte)(unsafe.Pointer(&keys[i])), sz)[ctr.keyOffs[i]+1:], vData[vOff[i+start]:vOff[i+start]+vLen[i+start]])
			ctr.keyOffs[i] += vLen[i+start] + 1
		}
	} else {
		for i := 0; i < n; i++ {
			if vec.Nsp.Np.Contains(uint64(i + start)) {
				*(*int8)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i])) = 1
				ctr.keyOffs[i]++
			} else {
				*(*int8)(unsafe.Add(unsafe.Pointer(&keys[i]), ctr.keyOffs[i])) = 0
				copy(unsafe.Slice((*byte)(unsafe.Pointer(&keys[i])), sz)[ctr.keyOffs[i]+1:], vData[vOff[i+start]:vOff[i+start]+vLen[i+start]])
				ctr.keyOffs[i] += vLen[i+start] + 1
			}
		}
	}
}

func fillGroupStr[T any](ctr *Container, vec *vector.Vector, n int, sz int, start int) {
	vs := vector.DecodeFixedCol[T](vec, sz)
	data := unsafe.Slice((*byte)(unsafe.Pointer(&vs[0])), cap(vs)*sz)[:len(vs)*sz]
	if !nulls.Any(vec.Nsp) {
		for i := 0; i < n; i++ {
			ctr.hstr.keys[i] = append(ctr.hstr.keys[i], byte(0))
			ctr.hstr.keys[i] = append(ctr.hstr.keys[i], data[(i+start)*sz:(i+start+1)*sz]...)
		}
	} else {
		for i := 0; i < n; i++ {
			if vec.Nsp.Np.Contains(uint64(i + start)) {
				ctr.hstr.keys[i] = append(ctr.hstr.keys[i], byte(1))
			} else {
				ctr.hstr.keys[i] = append(ctr.hstr.keys[i], byte(0))
				ctr.hstr.keys[i] = append(ctr.hstr.keys[i], data[(i+start)*sz:(i+start+1)*sz]...)
			}
		}
	}

}
