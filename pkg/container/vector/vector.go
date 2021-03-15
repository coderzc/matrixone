package vector

import (
	"bytes"
	"fmt"
	"matrixbase/pkg/container/nulls"
	"matrixbase/pkg/container/types"
	"matrixbase/pkg/encoding"
	"matrixbase/pkg/vectorize/shuffle"
	"matrixbase/pkg/vm/mempool"
	"matrixbase/pkg/vm/process"
	"reflect"
	"strconv"
	"unsafe"
)

func New(typ types.Type) *Vector {
	switch typ.Oid {
	case types.T_int8:
		return &Vector{
			Typ: typ,
			Col: []int8{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_int16:
		return &Vector{
			Typ: typ,
			Col: []int32{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_int32:
		return &Vector{
			Typ: typ,
			Col: []int64{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_int64:
		return &Vector{
			Typ: typ,
			Col: []int64{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_uint8:
		return &Vector{
			Typ: typ,
			Col: []uint8{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_uint16:
		return &Vector{
			Typ: typ,
			Col: []uint16{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_uint32:
		return &Vector{
			Typ: typ,
			Col: []uint32{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_uint64:
		return &Vector{
			Typ: typ,
			Col: []uint64{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_decimal:
		return &Vector{
			Typ: typ,
			Col: []types.Decimal{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_float32:
		return &Vector{
			Typ: typ,
			Col: []float32{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_float64:
		return &Vector{
			Typ: typ,
			Col: []float64{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_date:
		return &Vector{
			Typ: typ,
			Col: []types.Date{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_datetime:
		return &Vector{
			Typ: typ,
			Col: []types.Datetime{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_sel:
		return &Vector{
			Typ: typ,
			Col: []int64{},
			Nsp: &nulls.Nulls{},
		}
	case types.T_tuple:
		return &Vector{
			Typ: typ,
			Nsp: &nulls.Nulls{},
			Col: [][]interface{}{},
		}
	case types.T_char, types.T_varchar, types.T_json:
		return &Vector{
			Typ: typ,
			Col: &types.Bytes{},
			Nsp: &nulls.Nulls{},
		}
	}
	return nil
}

func (v *Vector) Reset() {
	switch v.Typ.Oid {
	case types.T_char, types.T_varchar, types.T_json:
		v.Col.(*types.Bytes).Reset()
	default:
		*(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&v.Col)) + uintptr(strconv.IntSize>>3))) = 0
	}
}

func (v *Vector) Free(p *process.Process) {
	if v.Data != nil {
		if p.Free(v.Data) {
			v.Data = nil
		}
	}
}

func (v *Vector) SetCol(col interface{}) {
	v.Col = col
}

func (v *Vector) Length() int {
	switch v.Typ.Oid {
	case types.T_char, types.T_varchar, types.T_json:
		return len(v.Col.(*types.Bytes).Offsets)
	default:
		hp := *(*reflect.SliceHeader)(unsafe.Pointer(&v.Col))
		return hp.Len
	}
}

func (v *Vector) Window(start, end int) *Vector {
	switch v.Typ.Oid {
	case types.T_char, types.T_varchar, types.T_json:
		return &Vector{
			Typ: v.Typ,
			Col: v.Col.(*types.Bytes).Window(start, end),
			Nsp: v.Nsp.Range(uint64(start), uint64(end)),
		}
	default:
		col := v.Col
		ptr := unsafe.Pointer(&col)
		data := *(*uintptr)(unsafe.Pointer(uintptr(ptr)))
		*(*uintptr)(unsafe.Pointer(uintptr(ptr))) = data + uintptr(v.Typ.Size)*uintptr(start)
		*(*int)(unsafe.Pointer(uintptr(ptr) + uintptr(strconv.IntSize>>3))) = end - start + 1
		return &Vector{
			Typ: v.Typ,
			Col: col,
			Nsp: v.Nsp.Range(uint64(start), uint64(end)),
		}
	}
}

func (v *Vector) Append(arg interface{}) error {
	switch v.Typ.Oid {
	case types.T_int8:
		v.Col = append(v.Col.([]int8), arg.([]int8)...)
	case types.T_int16:
		v.Col = append(v.Col.([]int16), arg.([]int16)...)
	case types.T_int32:
		v.Col = append(v.Col.([]int32), arg.([]int32)...)
	case types.T_int64:
		v.Col = append(v.Col.([]int64), arg.([]int64)...)
	case types.T_uint8:
		v.Col = append(v.Col.([]uint8), arg.([]uint8)...)
	case types.T_uint16:
		v.Col = append(v.Col.([]uint16), arg.([]uint16)...)
	case types.T_uint32:
		v.Col = append(v.Col.([]uint32), arg.([]uint32)...)
	case types.T_uint64:
		v.Col = append(v.Col.([]uint64), arg.([]uint64)...)
	case types.T_decimal:
		v.Col = append(v.Col.([]types.Decimal), arg.([]types.Decimal)...)
	case types.T_float32:
		v.Col = append(v.Col.([]float32), arg.([]float32)...)
	case types.T_float64:
		v.Col = append(v.Col.([]float64), arg.([]float64)...)
	case types.T_date:
		v.Col = append(v.Col.([]types.Date), arg.([]types.Date)...)
	case types.T_datetime:
		v.Col = append(v.Col.([]types.Datetime), arg.([]types.Datetime)...)
	case types.T_sel:
		v.Col = append(v.Col.([]int64), arg.([]int64)...)
	case types.T_tuple:
		v.Col = append(v.Col.([][]interface{}), arg.([][]interface{})...)
	case types.T_char, types.T_varchar, types.T_json:
		return v.Col.(*types.Bytes).Append(arg.([][]byte))
	}
	return nil
}

func (v *Vector) Shuffle(sels []int64) *Vector {
	switch v.Typ.Oid {
	case types.T_int8:
		vs := v.Col.([]int8)
		shuffle.I8Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_int16:
		vs := v.Col.([]int16)
		shuffle.I16Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_int32:
		vs := v.Col.([]int32)
		shuffle.I32Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_int64:
		vs := v.Col.([]int64)
		shuffle.I64Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_uint8:
		vs := v.Col.([]uint8)
		shuffle.Ui8Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_uint16:
		vs := v.Col.([]uint16)
		shuffle.Ui16Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_uint32:
		vs := v.Col.([]uint32)
		shuffle.Ui32Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_uint64:
		vs := v.Col.([]uint64)
		shuffle.Ui64Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_decimal:
		vs := v.Col.([]types.Decimal)
		shuffle.DecimalShuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_float32:
		vs := v.Col.([]float32)
		shuffle.Float32Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_float64:
		vs := v.Col.([]float64)
		shuffle.Float64Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_date:
		vs := v.Col.([]types.Date)
		shuffle.DateShuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_datetime:
		vs := v.Col.([]types.Datetime)
		shuffle.DatetimeShuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_sel:
		vs := v.Col.([]int64)
		shuffle.I64Shuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_tuple:
		vs := v.Col.([][]interface{})
		shuffle.TupleShuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	case types.T_char, types.T_varchar, types.T_json:
		vs := v.Col.(*types.Bytes)
		shuffle.SShuffle(vs, sels)
		v.Col = vs
		v.Nsp = v.Nsp.Filter(sels)
	}
	return nil
}

func (v *Vector) UnionOne(w *Vector, sel int64, proc *process.Process) error {
	return nil
}

func (v *Vector) Show() ([]byte, error) {
	var buf bytes.Buffer

	switch v.Typ.Oid {
	case types.T_int8:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeInt8Slice(v.Col.([]int8)))
		return buf.Bytes(), nil
	case types.T_int16:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeInt16Slice(v.Col.([]int16)))
		return buf.Bytes(), nil
	case types.T_int32:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeInt32Slice(v.Col.([]int32)))
		return buf.Bytes(), nil
	case types.T_int64:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeInt64Slice(v.Col.([]int64)))
		return buf.Bytes(), nil
	case types.T_uint8:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeUint8Slice(v.Col.([]uint8)))
		return buf.Bytes(), nil
	case types.T_uint16:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeUint16Slice(v.Col.([]uint16)))
		return buf.Bytes(), nil
	case types.T_uint32:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeUint32Slice(v.Col.([]uint32)))
		return buf.Bytes(), nil
	case types.T_uint64:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeUint64Slice(v.Col.([]uint64)))
		return buf.Bytes(), nil
	case types.T_decimal:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeDecimalSlice(v.Col.([]types.Decimal)))
		return buf.Bytes(), nil
	case types.T_float32:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeFloat32Slice(v.Col.([]float32)))
		return buf.Bytes(), nil
	case types.T_float64:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeFloat64Slice(v.Col.([]float64)))
		return buf.Bytes(), nil
	case types.T_date:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeDateSlice(v.Col.([]types.Date)))
		return buf.Bytes(), nil
	case types.T_datetime:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeDatetimeSlice(v.Col.([]types.Datetime)))
		return buf.Bytes(), nil
	case types.T_sel:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		buf.Write(encoding.EncodeInt64Slice(v.Col.([]int64)))
		return buf.Bytes(), nil
	case types.T_char, types.T_varchar, types.T_json:
		buf.Write(encoding.EncodeType(v.Typ))
		nb, err := v.Nsp.Show()
		if err != nil {
			return nil, err
		}
		buf.Write(encoding.EncodeUint32(uint32(len(nb))))
		if len(nb) > 0 {
			buf.Write(nb)
		}
		Col := v.Col.(*types.Bytes)
		cnt := int32(len(Col.Offsets))
		buf.Write(encoding.EncodeInt32(cnt))
		if cnt == 0 {
			return buf.Bytes(), nil
		}
		buf.Write(encoding.EncodeUint32Slice(Col.Lengths))
		buf.Write(Col.Data)
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupport encoding type %s", v.Typ.Oid)
	}
}

func (v *Vector) Read(data []byte) error {
	v.Data = data
	data = data[mempool.CountSize:]
	typ := encoding.DecodeType(data[:encoding.TypeSize])
	data = data[encoding.TypeSize:]
	switch typ.Oid {
	case types.T_int8:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeInt8Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeInt8Slice(data[size:])
		}
	case types.T_int16:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeInt16Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeInt16Slice(data[size:])
		}
	case types.T_int32:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeInt32Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeInt32Slice(data[size:])
		}
	case types.T_int64:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeInt64Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeInt64Slice(data[size:])
		}
	case types.T_uint8:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeUint8Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeUint8Slice(data[size:])
		}
	case types.T_uint16:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeUint16Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeUint16Slice(data[size:])
		}
	case types.T_uint32:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeUint32Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeUint32Slice(data[size:])
		}
	case types.T_uint64:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeUint64Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeUint64Slice(data[size:])
		}
	case types.T_decimal:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeDecimalSlice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeDecimalSlice(data[size:])
		}
	case types.T_float32:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeFloat32Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeFloat32Slice(data[size:])
		}
	case types.T_float64:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeFloat64Slice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeFloat64Slice(data[size:])
		}
	case types.T_date:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeDateSlice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeDateSlice(data[size:])
		}
	case types.T_datetime:
		size := encoding.DecodeUint32(data)
		if size == 0 {
			v.Col = encoding.DecodeDatetimeSlice(data[4:])
		} else {
			data = data[4:]
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			v.Col = encoding.DecodeDatetimeSlice(data[size:])
		}
	case types.T_char, types.T_varchar, types.T_json:
		Col := v.Col.(*types.Bytes)
		Col.Reset()
		size := encoding.DecodeUint32(data)
		data = data[4:]
		if size > 0 {
			if err := v.Nsp.Read(data[:size]); err != nil {
				return err
			}
			data = data[size:]
		}
		cnt := encoding.DecodeInt32(data)
		if cnt == 0 {
			break
		}
		data = data[4:]
		Col.Offsets = make([]uint32, cnt)
		Col.Lengths = encoding.DecodeUint32Slice(data[:4*cnt])
		Col.Data = data[4*cnt:]
		{
			o := uint32(0)
			for i, n := range Col.Lengths {
				Col.Offsets[i] = o
				o += n
			}
		}
	}
	return nil
}

func (v *Vector) String() string {
	switch v.Typ.Oid {
	case types.T_int8:
		col := v.Col.([]int8)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_int16:
		col := v.Col.([]int16)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_int32:
		col := v.Col.([]int32)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_int64:
		col := v.Col.([]int64)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_uint8:
		col := v.Col.([]uint8)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_uint16:
		col := v.Col.([]uint16)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_uint32:
		col := v.Col.([]uint32)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_uint64:
		col := v.Col.([]uint64)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_decimal:
		col := v.Col.([]types.Decimal)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_float32:
		col := v.Col.([]float32)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_float64:
		col := v.Col.([]float64)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_date:
		col := v.Col.([]types.Date)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_datetime:
		col := v.Col.([]types.Datetime)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_sel:
		col := v.Col.([]int64)
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_tuple:
		col := v.Col.([][]interface{})
		if len(col) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%v", col[0])
			}
		}
	case types.T_char, types.T_varchar, types.T_json:
		col := v.Col.(*types.Bytes)
		if len(col.Offsets) == 1 {
			if v.Nsp.Contains(0) {
				fmt.Print("null")
			} else {
				return fmt.Sprintf("%s", col.Data[:col.Lengths[0]])
			}
		}

	}
	return fmt.Sprintf("%v-%s", v.Col, v.Nsp)
}
