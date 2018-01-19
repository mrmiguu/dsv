package dsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ff = `settlement-id,settlement-start-date,settlement-end-date,deposit-date,total-amount,currency,transaction-type,order-id,merchant-order-id,adjustment-id,shipment-id,marketplace-name,shipment-fee-type,shipment-fee-amount,order-fee-type,order-fee-amount,fulfillment-id,posted-date,order-item-code,merchant-order-item-id,merchant-adjustment-item-id,sku,quantity-purchased,price-type,price-amount,item-related-fee-type,item-related-fee-amount,misc-fee-amount,other-fee-amount,other-fee-reason-description,promotion-id,promotion-type,promotion-amount,direct-payment-type,direct-payment-amount,other-amount
6538365141,2017-12-29T01:22:44+00:00,2018-01-12T01:22:44+00:00,2018-01-14T01:22:44+00:00,5026.51,USD,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
6538365141,,,,,,Order,114-1686460-5280236,16691606,,DftN2vcd9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,4.57658E+13,,,F15108105-HGR-42,1,,,,,,,,,,,,,
6538365141,,,,,,Order,114-1686460-5280236,16691606,,DftN2vcd9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,4.57658E+13,,,F15108105-HGR-42,,Principal,39.5,,,,,,,,,,,
6538365141,,,,,,Order,114-1686460-5280236,16691606,,DftN2vcd9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,4.57658E+13,,,F15108105-HGR-42,,Tax,3.75,,,,,,,,,,,
6538365141,,,,,,Order,114-1686460-5280236,16691606,,DftN2vcd9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,4.57658E+13,,,F15108105-HGR-42,,,,Commission,-5.93,,,,,,,,,
6538365141,,,,,,Order,114-1686460-5280236,16691606,,DftN2vcd9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,4.57658E+13,,,F15108105-HGR-42,,,,SalesTaxServiceFee,-0.11,,,,,,,,,
6538365141,,,,,,Order,111-6209521-8633824,16691925,,DzWd2mcB9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,3.12421E+13,,,DVF0000056-111-11,1,,,,,,,,,,,,,
6538365141,,,,,,Order,111-6209521-8633824,16691925,,DzWd2mcB9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,3.12421E+13,,,DVF0000056-111-11,,Principal,85,,,,,,,,,,,
6538365141,,,,,,Order,111-6209521-8633824,16691925,,DzWd2mcB9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,3.12421E+13,,,DVF0000056-111-11,,Shipping,10.27,,,,,,,,,,,
6538365141,,,,,,Order,111-6209521-8633824,16691925,,DzWd2mcB9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,3.12421E+13,,,DVF0000056-111-11,,,,Commission,-12.75,,,,,,,,,
6538365141,,,,,,Order,111-6209521-8633824,16691925,,DzWd2mcB9,Amazon.com,,,,,MFN,2018-01-11T22:58:20+00:00,3.12421E+13,,,DVF0000056-111-11,,,,ShippingHB,-1.54,,,,,,,,,`

var (
	println = fmt.Println

	kptr     = reflect.Ptr
	kstruct  = reflect.Struct
	kslice   = reflect.Slice
	kstring  = reflect.String
	kint     = reflect.Int
	kfloat64 = reflect.Float64
	ktime    = reflect.TypeOf(time.Time{}).Kind()
)

type value reflect.Value

func main() {

	var report struct {
		DepDate time.Time
		TotAmt  float64
		Cur     string

		Orders []struct {
			OrdID    string
			PostDate time.Time

			SKUs []struct {
				SKU string

				Principal float64
				Shipping  float64
				Tax       float64

				ItmRelFeeAmt float64
				MiscFeeAmt   float64
				OtherFeeAmt  float64
				PromoAmt     float64
				DirPayAmt    float64
				OtherAmt     float64
			}
		}
	}
	// println(report)

	decode(strings.NewReader(ff), &report)
	// println(report)

}

func decode(r io.Reader, v interface{}) error {
	v = deref(v)
	ff := csv.NewReader(r)

	for {
		row, err := ff.Read()
		if err == io.EOF {
			break
		}

		for _, fld := range fields(v) {
			switch fld.v.Kind() {
			case kslice:
			case kstring:
			case kint:
			case kfloat64:
			case ktime:
			default:
				panic("unknown type '" + fld.v.String() + "'")
			}
			// col, err := abbrev(fld.string, row...)
			// if err == nil {
			// 	fld.string
			// }
		}
	}

	return nil
}

func set(v reflect.Value, x interface{}) {
	v.Set(reflect.ValueOf(x))
}
func setstring(v reflect.Value, s string) {
	set(v, s)
}
func setint(v reflect.Value, i string) error {
	x, err := strconv.Atoi(i)
	if err != nil {
		return err
	}
	set(v, x)
	return nil
}
func setfloat(v reflect.Value, f string) error {
	x, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return err
	}
	set(v, x)
	return nil
}
func settime(v reflect.Value, t string) error {
	for _, timefmt := range [...]string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	} {
		if t, err := time.Parse(timefmt, t); err == nil {
			set(v, t)
			return nil
		}
	}
	return errors.New("unknown time format")
}

func gettype(v interface{}) string {
	return reflect.TypeOf(v).String()
}
func deref(v interface{}) reflect.Value {
	if reflect.TypeOf(v).Kind() != kptr {
		panic("not a pointer")
	}
	return reflect.ValueOf(v).Elem()
}

type field struct {
	string
	v reflect.Value
}

func fields(v interface{}) []field {
	if ValueOf(v).Kind() {
		panic("not a struct")
	}
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	flds := make([]field, rv.NumField())
	for i := range flds {
		flds[i] = field{rt.Field(i).Name, rv.Field(i)}
	}
	return flds
}

func abbrev(sub string, strs ...string) (int, error) {
	if len(sub) == 0 {
		return -1, errors.New("empty substring")
	}
	expr := `(?i)[^` + string(sub[0]) + `]*`
	for x, r := range sub {
		end := `.*`
		if xplus1 := x + 1; xplus1 < len(sub) {
			end = `[^` + string(sub[xplus1]) + `]*`
		}
		expr += string(r) + end
	}
	n := -1
	matches := []string{""}
	for i, str := range strs {
		l, L := len(str), len(matches[0])
		if l == 0 || !regexp.MustCompile(expr).MatchString(str) {
			continue
		} else if 0 < L && L < l {
			continue
		}
		if l == L {
			matches = append(matches, str)
		} else {
			matches = []string{str}
			n = i
		}
	}
	if len(matches[0]) == 0 {
		return -1, errors.New("no match for '" + sub + "'")
	} else if len(matches) > 1 {
		return -1, errors.New(sub + "=" + fmt.Sprint(matches) + " (too many matches)")
	}
	return n, nil
}
