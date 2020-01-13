package gwf

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type family struct {
	Name     string   `form:"name"`
	Num      int      `form:"num"`
	Children []string `form:"children"`
	WomonNum int      `form:"womon_num"`
	ManNum   int      `form:"man_num"`
}

func TestGetParams(t *testing.T) {
	targetUrl := "http://127.0.0.1/context?a=xxx&b=1234&c=3.2&d=true&q=apple&q=banana"
	u := url.Values{}
	u.Set("ff", "2222")
	u.Set("gg", "true")
	u.Set("dd", "2.32")
	u.Set("uu", "user")
	u.Set("ul", "l1")
	u.Add("ul", "l2")
	request, _ := http.NewRequest("POST", targetUrl, strings.NewReader(u.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c := newCtx(nil, request)
	a := c.ParamString("a")
	assert.Equal(t, a, "xxx")
	zStringD := c.ParamStringDefault("z", "yes")
	assert.Equal(t, zStringD, "yes")

	b := c.ParamInt("b")
	assert.Equal(t, b, 1234)
	zIntD := c.ParamIntDefault("z", 100)
	assert.Equal(t, zIntD, 100)

	bu := c.ParamUint("b")
	assert.EqualValues(t, bu, 1234)
	zUintD := c.ParamUintDefault("z", 100)
	assert.EqualValues(t, zUintD, 100)

	bInt8 := c.ParamInt8("b")
	assert.EqualValues(t, bInt8, 0)
	zInt8D := c.ParamInt8Default("z", 127)
	assert.EqualValues(t, zInt8D, 127)

	bUint8 := c.ParamUint8("b")
	assert.EqualValues(t, bUint8, 0)
	zUint8D := c.ParamUint8Default("z", 200)
	assert.EqualValues(t, zUint8D, 200)

	bInt16 := c.ParamInt16("b")
	assert.EqualValues(t, bInt16, 1234)
	zInt16D := c.ParamInt16Default("z", 300)
	assert.EqualValues(t, zInt16D, 300)

	bUint16 := c.ParamUint16("b")
	assert.EqualValues(t, bUint16, 1234)
	zUint16D := c.ParamUint16Default("z", 300)
	assert.EqualValues(t, zUint16D, 300)

	bInt32 := c.ParamInt32("b")
	assert.EqualValues(t, bInt32, 1234)
	zInt32D := c.ParamInt32Default("z", 400)
	assert.EqualValues(t, zInt32D, 400)

	bUint32 := c.ParamUint32("b")
	assert.EqualValues(t, bUint32, 1234)
	zUint32D := c.ParamUint32Default("z", 400)
	assert.EqualValues(t, zUint32D, 400)

	bInt64 := c.ParamInt64("b")
	assert.EqualValues(t, bInt64, 1234)
	zInt64D := c.ParamInt64Default("z", 500)
	assert.EqualValues(t, zInt64D, 500)

	bUint64 := c.ParamUint64("b")
	assert.EqualValues(t, bUint64, 1234)
	zUint64D := c.ParamUint64Default("z", 500)
	assert.EqualValues(t, zUint64D, 500)

	cc := c.ParamFloat32("c")
	assert.Equal(t, cc, float32(3.2))
	zFloat32D := c.ParamFloat32Default("z", 6.4)
	assert.Equal(t, zFloat32D, float32(6.4))

	ccF64 := c.ParamFloat64("c")
	assert.Equal(t, ccF64, float64(3.2))
	zFloat64D := c.ParamFloat64Default("z", 6.4)
	assert.Equal(t, zFloat64D, float64(6.4))

	d := c.ParamBool("d")
	assert.Equal(t, d, true)
	zBoolD := c.ParamBoolDefault("z", false)
	assert.Equal(t, zBoolD, false)

	q := c.ParamStringSlice("q")
	qLen := len(q)
	assert.Equal(t, qLen, 2)

	ffInt := c.FormInt("ff")
	assert.Equal(t, ffInt, 2222)

	ffIntD := c.FormIntDefault("zz", 99)
	assert.Equal(t, ffIntD, 99)

	ffInt8 := c.FormInt8("ff")
	assert.EqualValues(t, ffInt8, 0)

	ffInt8D := c.FormInt8Default("zz", 77)
	assert.EqualValues(t, ffInt8D, 77)

	ffInt16 := c.FormInt16("ff")
	assert.EqualValues(t, ffInt16, 2222)

	ffInt16D := c.FormInt16Default("zz", 88)
	assert.EqualValues(t, ffInt16D, 88)

	ffInt32 := c.FormInt32("ff")
	assert.EqualValues(t, ffInt32, 2222)

	ffInt32D := c.FormInt32Default("zz", 100)
	assert.EqualValues(t, ffInt32D, 100)

	ffInt64 := c.FormInt64("ff")
	assert.EqualValues(t, ffInt64, 2222)

	ffInt64D := c.FormInt64Default("zz", 200)
	assert.EqualValues(t, ffInt64D, 200)

	ffu := c.FormUint("ff")
	assert.EqualValues(t, ffu, 2222)
	ffud := c.FormUintDefault("zz", 89)
	assert.EqualValues(t, ffud, 89)

	ffu8 := c.FormUint8("ff")
	assert.EqualValues(t, ffu8, 0)
	ffu8d := c.FormUint8Default("zz", 88)
	assert.EqualValues(t, ffu8d, 88)

	ffu16 := c.FormUint16("ff")
	assert.EqualValues(t, ffu16, 2222)
	ffu16d := c.FormUint16Default("zz", 87)
	assert.EqualValues(t, ffu16d, 87)

	ffu32 := c.FormUint32("ff")
	assert.EqualValues(t, ffu32, 2222)
	ffu32d := c.FormUint32Default("zz", 86)
	assert.EqualValues(t, ffu32d, 86)

	ffu64 := c.FormUint64("ff")
	assert.EqualValues(t, ffu64, 2222)
	ffu64d := c.FormUint64Default("zz", 85)
	assert.EqualValues(t, ffu64d, 85)

	gg := c.FormBool("gg")
	assert.Equal(t, gg, true)
	ggd := c.FormBoolDefault("zz", false)
	assert.Equal(t, ggd, false)

	dd32 := c.FormFloat32("dd")
	assert.EqualValues(t, dd32, float32(2.32))
	dd32d := c.FormFloa32Default("zz", 1.23)
	assert.EqualValues(t, dd32d, float32(1.23))

	dd64 := c.FormFloat64("dd")
	assert.Equal(t, dd64, 2.32)
	dd64d := c.FormFloa64Default("zz", 1.23)
	assert.Equal(t, dd64d, 1.23)

	uu := c.FormString("uu")
	assert.Equal(t, uu, "user")
	uuD := c.FormStringDefault("zz", "way")
	assert.Equal(t, uuD, "way")

	ul := c.FormStringSlice("ul")
	assert.Equal(t, 2, len(ul))

	qa := c.QueryString("a")
	assert.Equal(t, qa, "xxx")
	qaD := c.QueryStringDefault("zz", "yzyz")
	assert.Equal(t, "yzyz", qaD)

	qb := c.QueryInt("b")
	assert.EqualValues(t, qb, 1234)
	qbD := c.QueryIntDefault("zz", 890)
	assert.EqualValues(t, qbD, 890)

	qb8 := c.QueryInt8("b")
	assert.EqualValues(t, qb8, 0)
	qb8D := c.QueryInt8Default("zz", 89)
	assert.EqualValues(t, qb8D, 89)

	qb16 := c.QueryInt16("b")
	assert.EqualValues(t, qb16, 1234)
	qb16D := c.QueryInt16Default("zz", 892)
	assert.EqualValues(t, qb16D, 892)

	qb32 := c.QueryInt32("b")
	assert.EqualValues(t, qb32, 1234)
	qb32D := c.QueryInt32Default("zz", 893)
	assert.EqualValues(t, qb32D, 893)

	qb64 := c.QueryInt64("b")
	assert.EqualValues(t, qb64, 1234)
	qb64D := c.QueryInt64Default("zz", 894)
	assert.EqualValues(t, qb64D, 894)

	qbu := c.QueryUint("b")
	assert.EqualValues(t, qbu, 1234)
	qbuD := c.QueryUintDefault("zz", 895)
	assert.EqualValues(t, qbuD, 895)

	qbu8 := c.QueryUint8("b")
	assert.EqualValues(t, qbu8, 0)
	qbu8D := c.QueryUint8Default("zz", 86)
	assert.EqualValues(t, qbu8D, 86)

	qbu16 := c.QueryUint16("b")
	assert.EqualValues(t, qbu16, 1234)
	qbu16D := c.QueryUint16Default("zz", 897)
	assert.EqualValues(t, qbu16D, 897)

	qbu32 := c.QueryUint32("b")
	assert.EqualValues(t, qbu32, 1234)
	qbu32D := c.QueryUint32Default("zz", 898)
	assert.EqualValues(t, qbu32D, 898)

	qbu64 := c.QueryUint64("b")
	assert.EqualValues(t, qbu64, 1234)
	qbu64D := c.QueryUint64Default("zz", 899)
	assert.EqualValues(t, qbu64D, 899)

	qc := c.QueryFloat32("c")
	assert.Equal(t, qc, float32(3.2))
	qcD := c.QueryFloa32Default("zz", 6.5)
	assert.Equal(t, qcD, float32(6.5))

	qc64 := c.QueryFloat64("c")
	assert.Equal(t, qc64, 3.2)
	qc64D := c.QueryFloa64Default("zz", 6.6)
	assert.Equal(t, qc64D, 6.6)

	qd := c.QueryBool("d")
	assert.Equal(t, qd, true)
	qdd := c.QueryBoolDefault("zz", false)
	assert.Equal(t, qdd, false)

	qqs := c.QueryStringSlice("q")
	assert.Equal(t, 2, len(qqs))

}

func TestGetRequestBody(t *testing.T) {
	u := url.Values{}
	u.Set("aa", "bb")
	u.Set("cc", "dd")
	r, _ := http.NewRequest("POST", "/post", strings.NewReader(u.Encode()))
	r.Header.Set("Content-Type", "application/json")
	c := newCtx(nil, r)
	b, err := c.GetReqeustBody()
	if err != nil {
		t.Fatalf("getReqeustBody err: %v", err)
	}
	assert.Equal(t, string(b), u.Encode())
}

func TestBind(t *testing.T) {
	u := url.Values{}
	u.Set("womon_num", "2")
	u.Set("man_num", "3")
	targetUrl := "http://127.0.0.1/family?name=family&num=5&children=lily&children=tom"
	r, _ := http.NewRequest("POST", targetUrl, strings.NewReader(u.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c := newCtx(nil, r)

	f1 := family{}
	err := c.BindQuery(&f1)
	if err != nil {
		t.Fatalf("BindQuery err: %v", err)
	}
	expectF1 := family{
		Name:     "family",
		Num:      5,
		Children: []string{"lily", "tom"},
		WomonNum: 0,
		ManNum:   0,
	}
	assert.Equal(t, f1, expectF1)

	f2 := family{}
	err = c.BindForm(&f2)
	if err != nil {
		t.Fatalf("BindForm err: %v", err)
	}
	expectF2 := family{
		Name:     "",
		Num:      0,
		Children: nil,
		WomonNum: 2,
		ManNum:   3,
	}
	assert.Equal(t, f2, expectF2)

	f3 := family{}
	err = c.BindParam(&f3)
	if err != nil {
		t.Fatalf("BindParam err: %v", err)
	}
	expectF3 := family{
		Name:     "family",
		Num:      5,
		Children: []string{"lily", "tom"},
		WomonNum: 2,
		ManNum:   3,
	}
	assert.Equal(t, f3, expectF3)
}

func TestHeader(t *testing.T) {
	r, _ := http.NewRequest("GET", "/header", nil)
	ctx := newCtx(nil, r)
	w := httptest.NewRecorder()
	ctx.Writer = NewResponseWriter(w, nil, nil, nil)
	ctx.Header("flag", "gopher")
	h1 := ctx.Writer.Header().Get("flag")
	assert.Equal(t, h1, "gopher")
}
