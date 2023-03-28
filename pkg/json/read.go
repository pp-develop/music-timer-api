package json

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
)

// JSONファイルから指定されたキーの値を文字列として取得する
func Read(key string) (string, error) {
    // ファイルの読み込み
    bytes, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    // JSONのパース
    var data map[string]interface{}
    err = json.Unmarshal(bytes, &data)
    if err != nil {
        return "", err
    }

    // 指定されたキーの値を取得
    value, ok := data[key]
    if !ok {
        return "", nil
    }

    // 値を文字列に変換して返す
    str, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("value of key '%s' is not a string", key)
    }

    return str, nil
}