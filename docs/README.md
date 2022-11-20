# 📗 Open API Template

# Usage
1. コンテナ実行
    ```
    $ docker-compose up -d
    ```
2. ファイルを更新
3. http://localhost:80 へ変更が反映されます  

    使用しているシェルスクリプト  
    https://github.com/kawamataryo/watch-files
# 📌TIPS
<!-- ## 編集中のOpenAPIのプレビュー
### VS Codeを使用

1. 拡張機能[Swagger Viewer](https://marketplace.visualstudio.com/items?itemName=Arjun.swagger-viewer)をインストール
2. ルートディレクトリのopenapi.ymlを開き、**Shift + Alt + P** -->

## ymlファイルを結合
[swagger-cli](https://github.com/APIDevTools/swagger-cli)のインストールが前提
```
/buildに結合したymlを出力
$ swagger-cli bundle -o ./build/openapi.yml -t yaml ./openapi.yml
```


## OpenAPIをRedocで出力
[redoc-cli](https://redocly.com/docs/redoc/deployment/cli/)のインストールが前提
```
$ redoc-cli build ./build/openapi.yml
```
