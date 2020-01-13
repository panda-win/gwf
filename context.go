package gwf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-playground/form"
)

// Context 是对某次请求上下文的抽象
type Context struct {
	app     *Application
	Request *http.Request
	Writer  *responseWriter

	//url参数列表
	URLParameters url.Values

	//url和POST/PUT/PATCH参数列表，url参数会被POST/PUT/PATCH的同名参数覆盖
	URLFormParameters url.Values

	//POST/PUT/PATCH参数列表
	FormParameters url.Values

	// HandlersChain的索引，用于控制handlers链执行流程
	index int8
	// 当前请求需要执行的所有handlers集合
	handlers HandlersChain

	// Keys key/value对，请求唯一，可以将参数贯串整个请求上下文
	Keys map[string]interface{}
	// 内部错误
	errInternal *Error
}

const abortIndex int8 = math.MaxInt8 / 2

var decoder *form.Decoder = form.NewDecoder()

func newCtx(app *Application, r *http.Request) *Context {
	c := &Context{
		app:     app,
		Request: r,
		index:   -1,
	}

	c.URLParameters = r.URL.Query()

	if r.Method == http.MethodPost {
		v := r.Header.Get("Content-Type")

		var parseMultipart bool = false

		if v != "" {
			d, _, err := mime.ParseMediaType(v)
			if err == nil && d == "multipart/form-data" {
				parseMultipart = true
			}
		}

		if parseMultipart {
			err := r.ParseMultipartForm(app.maxMultipartMemory)

			if err != nil {
				panic(fmt.Sprintf("ParseMultipartForm 失败, error:%s", err))
			}

		} else {
			err := r.ParseForm()
			if err != nil {
				//如果解析参数失败，后续流程无法进行，快速失败，日志中将会有记录
				panic(fmt.Sprintf("解析url参数和POST/PUT/PATCH上传参数失败, error:%s", err))
			}
		}

		c.URLFormParameters = c.Request.Form

	} else {
		c.URLFormParameters = c.URLParameters
	}

	c.FormParameters = c.Request.PostForm

	return c
}

// Next 循环执行hanlers链中的handler
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Status 写入响应code
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// IsAborted 当前Context终止返回true
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Abort 终止待执行的hanlers，不会终止正在执行中的handler
// 比如认证middleware执行失败时，不需要再执行之后的handlers
// 调用Abort来终止剩下的handlers的调用
func (c *Context) Abort() {
	c.index = abortIndex
}

// AbortWithStatus 写入status code并终止后续的handlers调用
func (c *Context) AbortWithStatus(code int) {
	c.Abort()
	c.Status(code)
	c.Writer.WriteHeaderNow()
}

// AbortWithStatusString 响应为string格式，并终止后续的handlers调用
func (c *Context) AbortWithStatusString(code int, data string) {
	c.Abort()
	c.String(code, data)
}

// AbortWithStatusJson 响应为json格式，并终止后续的handlers调用
func (c *Context) AbortWithStatusJson(code int, data interface{}) {
	c.Abort()
	c.Json(code, data)
}

// Set 在当前Context中保存用户自定义key/value
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get 取出当前Context中自定义key的value，存在时exists为true
func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

/**********请求参数相关方法 begin**********/
func intDefault(v string, defaultV int) int {
	if v == "" {
		return defaultV
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultV
	}
	return i
}

func int8Default(v string, defaultV int8) int8 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseInt(v, 10, 8)
	if err != nil {
		return defaultV
	}
	return int8(i)
}

func int16Default(v string, defaultV int16) int16 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseInt(v, 10, 16)
	if err != nil {
		return defaultV
	}
	return int16(i)
}

func int32Default(v string, defaultV int32) int32 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return defaultV
	}
	return int32(i)
}

func int64Default(v string, defaultV int64) int64 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultV
	}
	return i
}

func uintDefault(v string, defaultV uint) uint {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return defaultV
	}
	return uint(i)
}

func uint8Default(v string, defaultV uint8) uint8 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseUint(v, 10, 8)
	if err != nil {
		return defaultV
	}
	return uint8(i)
}

func uint16Default(v string, defaultV uint16) uint16 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		return defaultV
	}
	return uint16(i)
}

func uint32Default(v string, defaultV uint32) uint32 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return defaultV
	}
	return uint32(i)
}

func uint64Default(v string, defaultV uint64) uint64 {
	if v == "" {
		return defaultV
	}
	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return defaultV
	}
	return i
}

func float32Default(v string, defaultV float32) float32 {
	if v == "" {
		return defaultV
	}
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return defaultV
	}
	return float32(f)
}

func float64Default(v string, defaultV float64) float64 {

	if v == "" {
		return defaultV
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return defaultV
	}
	return f
}

func boolDefault(v string, defaultV bool) bool {
	if v == "" {
		return defaultV
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultV
	}
	return b
}

// ParamXXX和ParamXXXDefault函数可以取得url参数和POST、PUT、PATCH上传的参数
// 如果某个参数在url参数和POST、PUT、PATCH上传参数中都有，将取得POST、PUT、PATCH参数
// 如果一定要取得url参数，请使用QueryXXX和QueryXXXDefault方法
// 如果没有为key的参数名，或者值为空字符串，ParamXXX返回XXX类型的默认零值
// ParamXXXDefault返回传入的defaultV
func (c *Context) ParamString(key string) string {
	return c.URLFormParameters.Get(key)
}

func (c *Context) ParamStringDefault(key string, defaultV string) string {
	v := c.URLFormParameters.Get(key)
	if v == "" {
		return defaultV
	}

	return v
}

func (c *Context) ParamStringSlice(key string) []string {
	return c.URLFormParameters[key]
}

func (c *Context) ParamInt(key string) int {
	v := c.ParamString(key)
	return intDefault(v, 0)
}

func (c *Context) ParamIntDefault(key string, defaultV int) int {
	v := c.ParamString(key)
	return intDefault(v, defaultV)
}

func (c *Context) ParamInt8(key string) int8 {
	v := c.ParamString(key)
	return int8Default(v, int8(0))
}

func (c *Context) ParamInt8Default(key string, defaultV int8) int8 {
	v := c.ParamString(key)
	return int8Default(v, defaultV)
}

func (c *Context) ParamInt16(key string) int16 {
	v := c.ParamString(key)
	return int16Default(v, int16(0))
}

func (c *Context) ParamInt16Default(key string, defaultV int16) int16 {
	v := c.ParamString(key)
	return int16Default(v, defaultV)
}

func (c *Context) ParamInt32(key string) int32 {
	v := c.ParamString(key)
	return int32Default(v, int32(0))
}

func (c *Context) ParamInt32Default(key string, defaultV int32) int32 {
	v := c.ParamString(key)
	return int32Default(v, defaultV)
}

func (c *Context) ParamInt64(key string) int64 {
	v := c.ParamString(key)
	return int64Default(v, int64(0))
}

func (c *Context) ParamInt64Default(key string, defaultV int64) int64 {
	v := c.ParamString(key)
	return int64Default(v, defaultV)
}

func (c *Context) ParamUint(key string) uint {
	v := c.ParamString(key)
	return uintDefault(v, uint(0))
}

func (c *Context) ParamUintDefault(key string, defaultV uint) uint {
	v := c.ParamString(key)
	return uintDefault(v, defaultV)
}

func (c *Context) ParamUint8(key string) uint8 {
	v := c.ParamString(key)
	return uint8Default(v, uint8(0))
}

func (c *Context) ParamUint8Default(key string, defaultV uint8) uint8 {
	v := c.ParamString(key)
	return uint8Default(v, defaultV)
}

func (c *Context) ParamUint16(key string) uint16 {
	v := c.ParamString(key)
	return uint16Default(v, uint16(0))
}

func (c *Context) ParamUint16Default(key string, defaultV uint16) uint16 {
	v := c.ParamString(key)
	return uint16Default(v, defaultV)
}

func (c *Context) ParamUint32(key string) uint32 {
	v := c.ParamString(key)
	return uint32Default(v, uint32(0))
}

func (c *Context) ParamUint32Default(key string, defaultV uint32) uint32 {
	v := c.ParamString(key)
	return uint32Default(v, defaultV)
}

func (c *Context) ParamUint64(key string) uint64 {
	v := c.ParamString(key)
	return uint64Default(v, uint64(0))
}

func (c *Context) ParamUint64Default(key string, defaultV uint64) uint64 {
	v := c.ParamString(key)
	return uint64Default(v, defaultV)
}

func (c *Context) ParamFloat32(key string) float32 {
	v := c.ParamString(key)
	return float32Default(v, float32(0.0))
}

func (c *Context) ParamFloat32Default(key string, defaultV float32) float32 {
	v := c.ParamString(key)
	return float32Default(v, defaultV)
}

func (c *Context) ParamFloat64(key string) float64 {
	v := c.ParamString(key)
	return float64Default(v, 0.0)
}

func (c *Context) ParamFloat64Default(key string, defaultV float64) float64 {
	v := c.ParamString(key)
	return float64Default(v, defaultV)
}

func (c *Context) ParamBool(key string) bool {
	v := c.ParamString(key)
	return boolDefault(v, false)
}

func (c *Context) ParamBoolDefault(key string, defaultV bool) bool {
	v := c.ParamString(key)
	return boolDefault(v, defaultV)
}

func (c *Context) QueryString(key string) string {
	return c.URLParameters.Get(key)
}

func (c *Context) QueryStringDefault(key string, defaultV string) string {
	v := c.URLParameters.Get(key)
	if v == "" {
		return defaultV
	}
	return v
}

func (c *Context) QueryStringSlice(key string) []string {
	return c.URLParameters[key]
}

func (c *Context) QueryInt(key string) int {
	v := c.QueryString(key)
	return intDefault(v, 0)
}

func (c *Context) QueryIntDefault(key string, defaultV int) int {
	v := c.QueryString(key)
	return intDefault(v, defaultV)
}

func (c *Context) QueryInt8(key string) int8 {
	v := c.QueryString(key)
	return int8Default(v, 0)
}

func (c *Context) QueryInt8Default(key string, defaultV int8) int8 {
	v := c.QueryString(key)
	return int8Default(v, defaultV)
}

func (c *Context) QueryInt16(key string) int16 {
	v := c.QueryString(key)
	return int16Default(v, 0)
}

func (c *Context) QueryInt16Default(key string, defaultV int16) int16 {
	v := c.QueryString(key)
	return int16Default(v, defaultV)
}

func (c *Context) QueryInt32(key string) int32 {
	v := c.QueryString(key)
	return int32Default(v, 0)
}

func (c *Context) QueryInt32Default(key string, defaultV int32) int32 {
	v := c.QueryString(key)
	return int32Default(v, defaultV)
}

func (c *Context) QueryInt64(key string) int64 {
	v := c.QueryString(key)
	return int64Default(v, 0)
}

func (c *Context) QueryInt64Default(key string, defaultV int64) int64 {
	v := c.QueryString(key)
	return int64Default(v, defaultV)
}

func (c *Context) QueryUint(key string) uint {
	v := c.QueryString(key)
	return uintDefault(v, 0)
}

func (c *Context) QueryUintDefault(key string, defaultV uint) uint {
	v := c.QueryString(key)
	return uintDefault(v, defaultV)
}

func (c *Context) QueryUint8(key string) uint8 {
	v := c.QueryString(key)
	return uint8Default(v, 0)
}

func (c *Context) QueryUint8Default(key string, defaultV uint8) uint8 {
	v := c.QueryString(key)
	return uint8Default(v, defaultV)
}

func (c *Context) QueryUint16(key string) uint16 {
	v := c.QueryString(key)
	return uint16Default(v, 0)
}

func (c *Context) QueryUint16Default(key string, defaultV uint16) uint16 {
	v := c.QueryString(key)
	return uint16Default(v, defaultV)
}

func (c *Context) QueryUint32(key string) uint32 {
	v := c.QueryString(key)
	return uint32Default(v, 0)
}

func (c *Context) QueryUint32Default(key string, defaultV uint32) uint32 {
	v := c.QueryString(key)
	return uint32Default(v, defaultV)
}

func (c *Context) QueryUint64(key string) uint64 {
	v := c.QueryString(key)
	return uint64Default(v, 0)
}

func (c *Context) QueryUint64Default(key string, defaultV uint64) uint64 {
	v := c.QueryString(key)
	return uint64Default(v, defaultV)
}

func (c *Context) QueryFloat32(key string) float32 {
	v := c.QueryString(key)
	return float32Default(v, 0.0)
}

func (c *Context) QueryFloa32Default(key string, defaultV float32) float32 {
	v := c.QueryString(key)
	return float32Default(v, defaultV)
}

func (c *Context) QueryFloat64(key string) float64 {
	v := c.QueryString(key)
	return float64Default(v, 0.0)
}

func (c *Context) QueryFloa64Default(key string, defaultV float64) float64 {
	v := c.QueryString(key)
	return float64Default(v, defaultV)
}

func (c *Context) QueryBool(key string) bool {
	v := c.QueryString(key)
	return boolDefault(v, false)
}

func (c *Context) QueryBoolDefault(key string, defaultV bool) bool {
	v := c.QueryString(key)
	return boolDefault(v, defaultV)
}

func (c *Context) FormString(key string) string {
	return c.FormParameters.Get(key)
}

func (c *Context) FormStringDefault(key string, defaultV string) string {
	v := c.FormParameters.Get(key)
	if v == "" {
		return defaultV
	}
	return v
}

func (c *Context) FormStringSlice(key string) []string {
	return c.FormParameters[key]
}

func (c *Context) FormInt(key string) int {
	v := c.FormString(key)
	return intDefault(v, 0)
}

func (c *Context) FormIntDefault(key string, defaultV int) int {
	v := c.FormString(key)
	return intDefault(v, defaultV)
}

func (c *Context) FormInt8(key string) int8 {
	v := c.FormString(key)
	return int8Default(v, 0)
}

func (c *Context) FormInt8Default(key string, defaultV int8) int8 {
	v := c.FormString(key)
	return int8Default(v, defaultV)
}

func (c *Context) FormInt16(key string) int16 {
	v := c.FormString(key)
	return int16Default(v, 0)
}

func (c *Context) FormInt16Default(key string, defaultV int16) int16 {
	v := c.FormString(key)
	return int16Default(v, defaultV)
}

func (c *Context) FormInt32(key string) int32 {
	v := c.FormString(key)
	return int32Default(v, 0)
}

func (c *Context) FormInt32Default(key string, defaultV int32) int32 {
	v := c.FormString(key)
	return int32Default(v, defaultV)
}

func (c *Context) FormInt64(key string) int64 {
	v := c.FormString(key)
	return int64Default(v, 0)
}

func (c *Context) FormInt64Default(key string, defaultV int64) int64 {
	v := c.FormString(key)
	return int64Default(v, defaultV)
}

func (c *Context) FormUint(key string) uint {
	v := c.FormString(key)
	return uintDefault(v, 0)
}

func (c *Context) FormUintDefault(key string, defaultV uint) uint {
	v := c.FormString(key)
	return uintDefault(v, defaultV)
}

func (c *Context) FormUint8(key string) uint8 {
	v := c.FormString(key)
	return uint8Default(v, 0)
}

func (c *Context) FormUint8Default(key string, defaultV uint8) uint8 {
	v := c.FormString(key)
	return uint8Default(v, defaultV)
}

func (c *Context) FormUint16(key string) uint16 {
	v := c.FormString(key)
	return uint16Default(v, 0)
}

func (c *Context) FormUint16Default(key string, defaultV uint16) uint16 {
	v := c.FormString(key)
	return uint16Default(v, defaultV)
}

func (c *Context) FormUint32(key string) uint32 {
	v := c.FormString(key)
	return uint32Default(v, 0)
}

func (c *Context) FormUint32Default(key string, defaultV uint32) uint32 {
	v := c.FormString(key)
	return uint32Default(v, defaultV)
}

func (c *Context) FormUint64(key string) uint64 {
	v := c.FormString(key)
	return uint64Default(v, 0)
}

func (c *Context) FormUint64Default(key string, defaultV uint64) uint64 {
	v := c.FormString(key)
	return uint64Default(v, defaultV)
}

func (c *Context) FormFloat32(key string) float32 {
	v := c.FormString(key)
	return float32Default(v, 0.0)
}

func (c *Context) FormFloa32Default(key string, defaultV float32) float32 {
	v := c.FormString(key)
	return float32Default(v, defaultV)
}

func (c *Context) FormFloat64(key string) float64 {
	v := c.FormString(key)
	return float64Default(v, 0.0)
}

func (c *Context) FormFloa64Default(key string, defaultV float64) float64 {
	v := c.FormString(key)
	return float64Default(v, defaultV)
}

func (c *Context) FormBool(key string) bool {
	v := c.FormString(key)
	return boolDefault(v, false)
}

func (c *Context) FormBoolDefault(key string, defaultV bool) bool {
	v := c.FormString(key)
	return boolDefault(v, defaultV)
}

// MultipartFormParameters返回form的enctype="multipart/form-data"的POST/PUT/PATCH参数
func (c *Context) MultipartFormParameters() (url.Values, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(c.app.maxMultipartMemory); err != nil {
			return nil, err
		}
	}

	return c.Request.MultipartForm.Value, nil
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(c.app.maxMultipartMemory); err != nil {
			return nil, err
		}
	}
	_, fh, err := c.Request.FormFile(name)
	return fh, err
}

func (c *Context) GetReqeustBody() ([]byte, error) {

	return ioutil.ReadAll(c.Request.Body)
}

/**********请求参数相关函数 end**********/

/**********参数绑定到struct相关函数 begin**********/
// Bind 用于参数绑定，dst必须是struct的指针类型
func (c *Context) Bind(dst interface{}, src url.Values) error {
	err := decoder.Decode(dst, src)
	if err != nil {
		return err
	}
	return nil
}

// BindQuery 将url参数绑定到dst
func (c *Context) BindQuery(dst interface{}) error {
	err := c.Bind(dst, c.URLParameters)
	if err != nil {
		return err
	}
	return nil
}

// BindParam 将url参数和POST/PUT/PATCH参数绑定到dst
func (c *Context) BindParam(dst interface{}) error {
	err := c.Bind(dst, c.URLFormParameters)
	if err != nil {
		return err
	}
	return nil
}

// BindForm 将POST/PUT/PATCH参数绑定到dst
func (c *Context) BindForm(dst interface{}) error {
	err := c.Bind(dst, c.FormParameters)
	if err != nil {
		return err
	}
	return nil
}

// BindMultipartForm将POST/PUT/PATCH参数绑定到dst
// 与BindForm的区别是，此方法将绑定form的enctype="multipart/form-data"的参数
// 如果要获得上传的文件，请使用FormFile方法
func (c *Context) BindMultipartForm(dst interface{}) error {
	data, err := c.MultipartFormParameters()
	if err != nil {
		return err
	}
	err = c.Bind(dst, data)
	if err != nil {
		return err
	}
	return nil
}

/**********参数绑定到struct相关函数 end**********/

/**********输出相关函数 begin**********/
// Header 用于输出http协议header
func (c *Context) Header(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// Bytes 用于输出[]byte类型数据，并设置HTTP状态码为statusCode
// 调用方需要自己使用return控制程序流程
func (c *Context) Bytes(statusCode int, bytes []byte) {
	c.Status(statusCode)
	n, err := c.Writer.Write(bytes)
	if err != nil {
		panic(fmt.Sprintf("错误 err:%s byte sent:%d", err, n))
	}
}

// String 用于输出string类型数据，并设置HTTP状态码为statusCode
// 调用方需要自己使用return控制程序流程
func (c *Context) String(statusCode int, data string) {
	c.Writer.Header().Add("Content-Type", "text/plain; charset=UTF-8")
	c.Bytes(statusCode, []byte(data))
}

// Render 将context数据注入到模板中渲染
func (c *Context) Render(code int, layoutName, tmplName string, data map[string]interface{}) {
	data["_ctx"] = c
	Render(c.Writer, code, layoutName, tmplName, data)
}

// RenderAdmin 渲染后台模板
func (c *Context) RenderAdmin(layoutName, tmplName string, data map[string]interface{}) {
	menuList := LoadMenuList(layoutName, nil)
	menuList.SetActive(c.Request.URL.Path)
	c.Render(http.StatusOK, layoutName, tmplName, data)
}

// RenderAdminDefaultLayout 渲染后台默认模板
func (c *Context) RenderAdminDefaultLayout(tmplName string, data map[string]interface{}) {
	adminDefaultLayoutName := "admin/default"
	c.RenderAdmin(adminDefaultLayoutName, tmplName, data)
}

// Json 将data输出，data一般是一个struct
// 调用方需要自己使用return控制程序流程
func (c *Context) Json(code int, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("错误 err:%s", err))
	}
	c.Writer.Header().Add("Content-Type", "application/json; charset=UTF-8")
	c.Bytes(code, b)
}

func (c *Context) Redirect301(location string) {

	http.Redirect(c.Writer, c.Request, location, 301)
}

func (c *Context) Redirect302(location string) {

	http.Redirect(c.Writer, c.Request, location, 302)
}

/**********输出相关函数 end**********/
