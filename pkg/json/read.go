package json

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
)

// JSONファイルの中身をキーで指定して配列型の中身を取得する
func Read(key string) ([]interface{}, error) {
    // ファイルの読み込み
    bytes, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    // JSONのパース
    var data map[string]interface{}
    err = json.Unmarshal(bytes, &data)
    if err != nil {
        return nil, err
    }

    // 配列型のキーを取得
    value, ok := data[key]
    if !ok {
        return nil, fmt.Errorf("Key %s not found", key)
    }

    // 配列にキャスト
    array, ok := value.([]interface{})
    if !ok {
        return nil, fmt.Errorf("Key %s is not an array", key)
    }

    return array, nil
}