package global

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	Pa "path"
)

func Init() {
	// 1. 创建必要目录
	dirs := []string{
		"lang",
		GlobalConfig.LogDir,
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}
	}

	InitLogger()
	Log.Info("日志系统初始化完成")

	InitLang()
	Log.Infof("语言系统初始化完成，当前语言: %s", useLangPath)
}

func GetLangTextMap() map[string]string {
	langPack, err := GetLangPack()
	if err != nil {
		Log.Warnf("获取语言包失败: %v", err)
		return make(map[string]string)
	}
	return langPack.Textmap
}

func GetLangPack() (*LanguagePack, error) {
	langPath := Pa.Join(pathLang, useLangPath)

	if cached, ok := langPackCache[langPath]; ok {
		return cached, nil
	}

	if _, err := os.Stat(langPath); err == nil {
		pack, err := tryLoadLangPack(langPath)
		if err == nil {
			langPackCache[langPath] = pack
		}
		return pack, err
	}

	embedPath := Pa.Join("lang", useLangPath)
	pack, err := tryLoadLangPackFromEmbed(embedPath)
	if err == nil {
		langPackCache[langPath] = pack
	}
	return pack, err
}

func tryLoadLangPack(langPath string) (*LanguagePack, error) {
	infoPath := Pa.Join(langPath, "info.json")
	infoData, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	var langInfo LanguageInfo
	if err := json.Unmarshal(infoData, &langInfo); err != nil {
		return nil, err
	}

	textmapPath := Pa.Join(langPath, "textmap.json")
	textmapData, err := os.ReadFile(textmapPath)
	if err != nil {
		return nil, fmt.Errorf("读取 textmap.json 失败: %v", err)
	}

	var textmap map[string]string
	if err := json.Unmarshal(textmapData, &textmap); err != nil {
		return nil, fmt.Errorf("解析 textmap.json 失败: %v", err)
	}

	return &LanguagePack{
		LanguageInfo: langInfo,
		Textmap:      textmap,
	}, nil
}

func tryLoadLangPackFromEmbed(langPath string) (*LanguagePack, error) {
	if LangFS == nil {
		return nil, fmt.Errorf("嵌入的文件系统未初始化")
	}

	infoPath := Pa.Join(langPath, "info.json")
	infoData, err := fs.ReadFile(LangFS, infoPath)
	if err != nil {
		Log.Warnf("读取嵌入 info.json 失败: %v", err)
		return nil, err
	}

	var langInfo LanguageInfo
	if err := json.Unmarshal(infoData, &langInfo); err != nil {
		Log.Warnf("解析嵌入 info.json 失败: %v", err)
		return nil, err
	}

	textmapPath := Pa.Join(langPath, "textmap.json")
	textmapData, err := fs.ReadFile(LangFS, textmapPath)
	if err != nil {
		return nil, fmt.Errorf("读取嵌入 textmap.json 失败: %v", err)
	}

	var textmap map[string]string
	if err := json.Unmarshal(textmapData, &textmap); err != nil {
		return nil, fmt.Errorf("解析嵌入 textmap.json 失败: %v", err)
	}

	return &LanguagePack{
		LanguageInfo: langInfo,
		Textmap:      textmap,
	}, nil
}
