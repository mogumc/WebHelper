package global

import (
	"encoding/json"
	"io/fs"
	"os"
	Pa "path"
	"path/filepath"
)

var (
	pathLang      = "lang/"
	useLangPath   = "default"
	allLangInfo   = []LanguageInfo{}
	LangFS        fs.FS
	langPackCache = map[string]*LanguagePack{}
	langCodeToDir = map[string]string{}
)

type LanguageInfo struct {
	LanguageName        string `json:"language_name"`
	LanguageCode        string `json:"language_code"`
	TextmapPath         string `json:"textmap_path"`
	TranslationProgress string `json:"translation_progress"`
	Translator          string `json:"translator"`
	LastUpdated         string `json:"last_updated"`
	Version             string `json:"version"`
}

type LanguagePack struct {
	LanguageInfo
	Textmap map[string]string `json:"textmap"`
}

func InitLang() {
	useLang := GlobalConfig.Language
	allLangInfo = []LanguageInfo{}
	langCodeToDir = map[string]string{}

	// 扫描文件系统语言目录
	scanFileSystemLangs(langCodeToDir)
	// 扫描嵌入文件系统语言
	scanEmbeddedLangs(langCodeToDir)

	// 设置当前语言路径
	if dir, ok := langCodeToDir[useLang]; ok {
		Log.Infof("找到匹配语言: %s -> %s", useLang, dir)
		useLangPath = dir
	} else {
		Log.Infof("未找到匹配语言 %s，使用默认语言", useLang)
		useLangPath = "default"
	}
}

func scanFileSystemLangs(langDirMap map[string]string) {
	err := filepath.WalkDir(pathLang, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			Log.Warnf("读取语言文件夹失败: %v", err)
			return nil
		}
		if !d.IsDir() || path == pathLang {
			return nil
		}
		return loadLangInfoFromFile(path, langDirMap, "[文件系统]")
	})
	if err != nil {
		Log.Warnf("遍历文件系统 lang 目录出错: %v", err)
	}
}

func scanEmbeddedLangs(langDirMap map[string]string) {
	if LangFS == nil {
		Log.Warnf("读取语言文件夹失败: %v", "LangFS 为空")
		return
	}

	entries, err := fs.ReadDir(LangFS, "lang")
	if err != nil {
		Log.Warnf("读取语言文件夹失败: %v", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		loadLangInfoFromEmbed(entry.Name(), langDirMap, "[嵌入FS]")
	}
}

func loadLangInfoFromFile(path string, langDirMap map[string]string, source string) error {
	infoPath := Pa.Join(path, "info.json")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		Log.Warnf("读取语言文件夹失败: %v", err)
		return nil
	}
	return parseAndAddLangInfo(data, filepath.Base(path), langDirMap, source)
}

func loadLangInfoFromEmbed(path string, langDirMap map[string]string, source string) {
	infoPath := Pa.Join("lang", path, "info.json")
	data, err := fs.ReadFile(LangFS, infoPath)
	if err != nil {
		Log.Warnf("读取语言文件夹失败: %v", err)
		return
	}
	parseAndAddLangInfo(data, filepath.Base(path), langDirMap, source)
}

func parseAndAddLangInfo(data []byte, dirName string, langDirMap map[string]string, source string) error {
	var info LanguageInfo
	if err := json.Unmarshal(data, &info); err != nil {
		Log.Warnf("解析语言文件夹失败: %v", err)
		return nil
	}
	if info.LanguageCode == "" || info.TextmapPath == "" {
		Log.Warnf("语言信息不完整: %s", dirName)
		return nil
	}
	if !containsLang(allLangInfo, info.LanguageCode) {
		allLangInfo = append(allLangInfo, info)
		langDirMap[info.LanguageCode] = dirName
		Log.Debugf("%s 识别到语言: %s (%s) -> %s", source, info.LanguageName, info.LanguageCode, dirName)
	}
	return nil
}

func containsLang(slice []LanguageInfo, code string) bool {
	for _, s := range slice {
		if s.LanguageCode == code {
			return true
		}
	}
	return false
}

func ClearLangCache() {
	langPackCache = map[string]*LanguagePack{}
}

func GetLangInfoList() []LanguageInfo {
	return allLangInfo
}

func UpdateCurrentLangPath() {
	useLang := GlobalConfig.Language
	if dir, ok := langCodeToDir[useLang]; ok {
		Log.Infof("更新语言路径: %s -> %s", useLang, dir)
		useLangPath = dir
	} else {
		Log.Infof("未找到匹配语言 %s，使用默认语言", useLang)
		useLangPath = "default"
	}
}
