package main

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
    "log"
    "fmt"
    "errors"
    "github.com/deanishe/awgo"
    "golang.org/x/text/unicode/norm"
)

type Result struct {
    Successful  bool    `json:"successful"`
    Text        string  `json:"text"`
    Translated  string  `json:"translated_text"`
}

var (
    endpoint    = "https://api.codic.jp/v1/engine/translate.json"
    wf          *aw.Workflow
)

func init() {
    wf = aw.New()
}

func run() {

    var text string
    var token string
    var project_id string
    var acronym_style string
    var casing string

    if args := wf.Args(); len(args) > 1 {
        text            = norm.NFC.String(args[0])
        token           = args[1]
        project_id      = args[2]
        acronym_style   = args[3]
        casing          = args[4]
    }

    log.Printf("text=%s, token=%s, project_id=%s, acronym_style=%s, casing=%s", text, token, project_id, acronym_style, casing)

    if token == "" {
        wf.FatalError(errors.New("Token is empty."))
    }

    if text == "" {
        // AlfredWorkflowの設定を'Arguments Required'にしているので通常ありえない
        wf.WarnEmpty("No results", "Try a different query?")
    }

    // クエリパラメータの設定
    values := url.Values{}
    values.Set("text", text)
    values.Add("token", token)
    if project_id != "" {
        values.Add("project_id", project_id)
    }
    if acronym_style != "" {
        values.Add("acronym_style", acronym_style)
    }
    if casing != "" {
        values.Add("casing", casing)
    }

    // Encode状態だとAPIが失敗する
    str, _ := url.QueryUnescape(values.Encode())
    req, err := http.NewRequest("POST", endpoint, strings.NewReader(str))
    if err != nil {
        wf.FatalError(err)
        log.Fatal(err)
    }

    // 認証トークンの設定
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
    // Content-Type 設定
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    // req.URL.RawQuery = values.Encode()

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        wf.FatalError(err)
        log.Fatal(err)
    }

    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        wf.FatalError(fmt.Errorf("Error StatusCode: %d", res.StatusCode))
        log.Fatal(res)
    }

    byteArray, err := ioutil.ReadAll(res.Body)
    if err != nil {
        wf.FatalError(err)
        log.Fatal(err)
    }

    jsonBytes := ([]byte)(byteArray)
    var data []Result

    log.Printf(string(jsonBytes))

    if err := json.Unmarshal(jsonBytes, &data); err != nil {
        wf.FatalError(err)
        log.Fatal(err)
    }

    for _, p := range data {
        wf.NewItem(p.Translated).
            Arg(p.Translated).
            UID(p.Translated).
            Valid(true)
        log.Printf(p.Translated)
    }

    wf.WarnEmpty("No results", "Try a different query?")

    wf.SendFeedback()
}

func main() {
    wf.Run(run)
}
