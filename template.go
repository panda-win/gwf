package gwf

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/panda-win/gwf/menu"
)

var log = log.New(os.Stdout, "gwf: ", log.Lshortfile)

// template 名称是由layout和tmpl名称组成的,详见getTemplateName
type templateName string

var enableDebug = false

var customFuncMap = template.FuncMap{}
var templateLoaded = false
var layouts map[string]*template.Template
var layoutMenuList map[string]menu.MenuList
var tmpls map[string]string

// 最终模板的缓存，加快速度
var cachedTemplate map[templateName]*template.Template

var updateTemplateMutex sync.Mutex

var delimiterLeft = "{{"
var delimiterRight = "}}"

const (
	templateFileDir = "template"
	layoutSuffix    = ".layout"
	tmplSuffix      = ".tmpl"
)

// listAllFileInDir将基于dir，找到所有文件名称（不包含目录），并且可以通过ext(后缀)做筛选
func listAllFileInDir(dir, ext string) (filepaths []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printff("list file failed! dir:%s err:%v", dir, err)
		}

		if info.IsDir() || !strings.HasSuffix(path, ext) {
			return nil
		}
		filepaths = append(filepaths, path)
		return nil
	})

	return
}

func getTemplateDirPath() string {
	return fmt.Sprintf("%s/%s/", rootPath, templateFileDir)
}

func getLayoutFilepath(layoutName string) string {
	return fmt.Sprintf("%slayout/%s.layout", getTemplateDirPath(), layoutName)
}

func getLayoutMenuDefaultFilepath(layoutName string) string {
	return fmt.Sprintf("%slayout/%s.menu", getTemplateDirPath(), layoutName)
}

func getTmplFilepath(tmplName string) string {
	return fmt.Sprintf("%stmpl/%s.tmpl", getTemplateDirPath(), tmplName)
}

// EnableDebug开启debug模式，debug模式下，可以随便修改layout/tmpl文件定义。
// 生产环境下，为了性能，请一定不要调用此函数
func EnableDebug() {
	enableDebug = true
}

// LoadMenuList为名称为layoutName的layout加载目录，第二个参数传入nil，将找layout文件同名的.menu文件加载
// .menu文件格式为json格式
func LoadMenuList(layoutName string, jsonData []byte) menu.MenuList {
	if jsonData == nil {
		content, err := ioutil.ReadFile(getLayoutMenuDefaultFilepath(layoutName))
		if err != nil {
			log.Printf("加载layout模板目录数据失败! filepath: %s err:%v", getLayoutMenuDefaultFilepath(layoutName), err)
			return nil
		}
		jsonData = content
	}
	layoutMenuList = make(map[string]menu.MenuList)

	var menuList menu.MenuList
	err := json.Unmarshal(jsonData, &menuList)
	if err != nil {
		panic(fmt.Sprintf("加载目录失败"))
	}
	layoutMenuList[layoutName] = menuList
	return menuList
}

func LoadTemplate() {
	updateTemplateMutex.Lock()
	defer updateTemplateMutex.Unlock()

	if templateLoaded || enableDebug {
		return
	}

	layouts = make(map[string]*template.Template)
	layoutMenuList = make(map[string]menu.MenuList)
	tmpls = make(map[string]string)
	cachedTemplate = make(map[templateName]*template.Template)

	root := getTemplateDirPath()
	layoutRoot := root + "layout/"
	tmplRoot := root + "tmpl/"

	//加载layout
	layoutFilepaths, err := listAllFileInDir(layoutRoot, layoutSuffix)
	if err != nil {
		panic(err)
	}
	for _, p := range layoutFilepaths {
		layoutName := strings.TrimSuffix(strings.TrimPrefix(p, layoutRoot), layoutSuffix)
		content, err := ioutil.ReadFile(p)
		if err != nil {
			log.Printf("[ERROR] 加载layout模板内容失败! filepath: %s err:%v", p, err)
			continue
		}
		tmpl, err := template.New(layoutName).Delims(delimiterLeft, delimiterRight).Funcs(customFuncMap).Parse(string(content))
		if err != nil {
			log.Printf("[ERROR] 加载layout模板失败，请修复模板文件内容! file: template/%s.layout err:%v", layoutName, err)
			continue
		}
		layouts[layoutName] = tmpl
	}

	//加载tmpl
	tmplFilepaths, err := listAllFileInDir(tmplRoot, tmplSuffix)
	if err != nil {
		panic(err)
	}
	for _, p := range tmplFilepaths {
		tmplName := strings.TrimSuffix(strings.TrimPrefix(p, tmplRoot), tmplSuffix)
		content, err := ioutil.ReadFile(p)
		if err != nil {
			log.Printf("[ERROR] 加载tmpl模板内容失败! filepath: %s err:%v", p, err)
			continue
		}
		tmpls[tmplName] = string(content)
	}

	templateLoaded = true
}

// ReloadTemplate用来重新加载模板，重新加载后，缓存的模板会失效
func ReloadTemplate() {
	LoadTemplate()
}

// SetFuncMap设定模板中的自定义函数，此方法要在LoadAdminTemplate之前调用才有效
// 比如自定义一个时间格式化函数:
//  func formatAsDate(t time.Time) string {
//		year, month, day := t.Date()
//		return fmt.Sprintf("%d%02d/%02d", year, month, day)
//	}
//
// 可以使用如下代码设定一个可以在模板文件中调用的formatAsDate函数:
//  template.SetFuncMap(map[string]interface{}{
//		"formatAsDate": formatAsDate,
//	})
//
// 在模板中可以使用自定义函数:
//  Date: {{.now | formatAsDate}}
func SetFuncMap(funcMap map[string]interface{}) {
	for k, f := range funcMap {
		if _, ok := customFuncMap[k]; ok {
			panic(fmt.Sprintf("SetFuncMap失败! %s已经预定义了，请不要重复定义", k))
		}
		customFuncMap[k] = f
	}
}

// SetDelimiters用来自定义定界符, 默认定界符是SetDelimiters("{{", "}}")
func SetDelimiters(left, right string) {
	delimiterLeft = left
	delimiterRight = right
}

var htmlContentType = []string{"text/html; charset=utf-8"}

func getTemplateName(layoutName, tmplName string) templateName {
	return templateName(fmt.Sprintf("%s:%s", layoutName, tmplName))
}

func RenderDebug(w http.ResponseWriter, code int, layoutName, tmplName string, data map[string]interface{}) {

	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = htmlContentType
	}
	var content []byte
	var err error
	var layoutTemplate *template.Template
	content, err = ioutil.ReadFile(getLayoutFilepath(layoutName))
	if err != nil {
		errMsg := fmt.Sprintf("加载layout模板内容失败! name:%s filepath: %s err:%v", layoutName, getLayoutFilepath(layoutName), err)
		log.Printf("[ERROR] " + errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	layoutTemplate, err = template.New(layoutName).Delims(delimiterLeft, delimiterRight).Funcs(customFuncMap).Parse(string(content))
	if err != nil {
		errMsg := fmt.Sprintf("加载layout模板失败，请修复模板文件内容! file: template/%s.layout err:%v", layoutName, err)
		log.Printf("[ERROR] " + errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	content, err = ioutil.ReadFile(getTmplFilepath(tmplName))
	if err != nil {
		errMsg := fmt.Sprintf("加载tmpl模板内容失败! name:%s filepath: %s err:%v", tmplName, getTmplFilepath(tmplName), err)
		log.Printf("[ERROR] " + errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var tmplContent string = string(content)
	var combinedTemplate *template.Template
	if combinedTemplate, err = layoutTemplate.Parse(tmplContent); err != nil {
		errMsg := fmt.Sprintf("combiledTemplate parse tmpl content failed! layoutName:%s tmplName:%s err:%v", layoutName, tmplName, err)
		log.Printf("[ERROR] " + errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	err = combinedTemplate.Execute(w, getRenderData(layoutName, data))
	if err != nil {
		errMsg := fmt.Sprintf("render failed! err:%v", err)
		log.Printf("[ERROR] " + errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func Render(w http.ResponseWriter, code int, layoutName, tmplName string, data map[string]interface{}) {
	updateTemplateMutex.Lock()
	defer updateTemplateMutex.Unlock()

	//debug模式
	if enableDebug {
		RenderDebug(w, code, layoutName, tmplName, data)
		return
	}

	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = htmlContentType
	}
	var layoutTemplate *template.Template
	var tmplContent string
	var combinedTemplate *template.Template
	var ok bool
	var err error
	if layoutTemplate, ok = layouts[layoutName]; !ok {
		errMsg := fmt.Sprintf("布局文件不存在! 名称:%s 路径:%s", layoutName, getLayoutFilepath(layoutName))
		log.Printf("[ERROR] " + errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if tmplContent, ok = tmpls[tmplName]; !ok {
		errMsg := fmt.Sprintf("模板文件不存在! 名称:%s 路径:%s", tmplName, getTmplFilepath(tmplName))
		log.Error(errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	combinedTemplateName := getTemplateName(layoutName, tmplName)
	if combinedTemplate, ok = cachedTemplate[combinedTemplateName]; !ok {
		//还没有缓存此模板
		if combinedTemplate, err = layoutTemplate.Clone(); err != nil {
			errMsg := fmt.Sprintf("clone layout template failed! layoutName:%s err:%v", layoutName, err)
			log.Error(errMsg)
			w.Write([]byte(errMsg))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if combinedTemplate, err = combinedTemplate.Parse(tmplContent); err != nil {
			errMsg := fmt.Sprintf("combiledTemplate parse tmpl content failed! layoutName:%s tmplName:%s err:%v", layoutName, tmplName, err)
			log.Error(errMsg)
			w.Write([]byte(errMsg))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//将最终模板缓存起来
		cachedTemplate[combinedTemplateName] = combinedTemplate
	}

	w.WriteHeader(code)
	err = combinedTemplate.Execute(w, getRenderData(layoutName, data))
	if err != nil {
		errMsg := fmt.Sprintf("render failed! template name:%s err:%v", combinedTemplateName, err)
		log.Error(errMsg)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getRenderData(layoutName string, data map[string]interface{}) map[string]interface{} {
	if menuList, ok := layoutMenuList[layoutName]; ok {
		if data == nil {
			data = map[string]interface{}{"_menuList": menuList}
		} else {
			data["_menuList"] = menuList
		}
	}
	return data
}
