/*
はい、もちろんです。ポップアップ内でユーザー名検索を行い、その結果から `userno` を選択してメイン画面のデータを更新する、という一連の処理も、前回のAjax通信の仕組みを応用することで実現可能です。

基本的な考え方は、**ポップアップ内でさらにAjax通信を行い、ユーザー名検索の結果を動的に表示する**というものです。

以下に、具体的な実装方法の概要とコード例を示します。

---

### 実装の全体像

1.  **Goハンドラーの追加**:
  - ユーザー名検索リクエストを受け付け、`userno` と `username` のリストをJSON形式で返す新しいエンドポイント（例: `/search-users`）を追加します。

2.  **HTML/CSSの変更**:
  - ポップアップ内に、ユーザー名入力用のテキストフィールド、検索ボタン、検索結果を表示する領域（例: `<ul>` リスト）を追加します。

3.  **JavaScriptの変更**:
  - ポップアップ内の「検索」ボタンが押されたら、入力されたユーザー名をサーバーの `/search-users` エンドポイントにAjaxリクエストで送信します。
  - サーバーからのJSONレスポンスを受け取り、ポップアップ内の検索結果表示領域をJavaScriptで動的に構築します。
  - ユーザーが検索結果リストからユーザー名を選択（クリック）したら、その `userno` を取得し、前回の「再表示」ボタンと同様に、メイン画面のデータ更新Ajaxリクエストを送信します。

---

### 具体的な実装例

#### 1. Goのハンドラーとテンプレートの準備

**`main.go`**

```go

v0.0.0  gemini-2.5-flashによるのコード生成(id=2308)
v0.0.1  .gitignoreを追加
v0.1.0  複数ダイアログに対応する（id=2321）
*/
package main

import (
	"crypto/tls"
	"encoding/json" // JSONを扱うために追加
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings" // 部分一致検索のために追加
	"time"
)

// 検索結果のデータ構造 (例)
type Data struct {
	ID    int
	Name  string
	Value string
}

// ユーザー検索結果のデータ構造
type User struct {
	UserNo   string `json:"userno"`
	UserName string `json:"username"`
}

// テンプレートに渡すデータ構造
type PageData struct {
	EventID string
	UserNo  string // 現在表示されているuserno
	Data    []Data
	IsAjax  bool
}

var tmpl *template.Template

func init() {
	// テンプレートファイルのパスは適宜調整してください
	tmpl = template.Must(template.ParseFiles("templates/index.html", "templates/data_table.html"))
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/search-users", searchUsersHandler) // 新しいハンドラーを追加
	// log.Fatal(http.ListenAndServe(":8080", nil))

	// HTTPサーバーを設定
	server := &http.Server{
		Addr:      ":" + "8000",
		TLSConfig: &tls.Config{
			// 必要に応じてTLS設定を追加
		},
		// Handler:      http.DefaultServeMux, // ここでハンドラを指定
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 90 * time.Second,
	}

	// サーバーをTLSで起動
	err := server.ListenAndServeTLS("/home/chouette/.ssh/cert.pem", "/home/chouette/.ssh/key.pem")
	if err != nil {
		log.Println("Server error:", err)
	}
}

// メイン画面の表示とデータ更新を行うハンドラー (前回とほぼ同じ)
func handler(w http.ResponseWriter, r *http.Request) {
	eventid := r.FormValue("eventid")
	userno := r.FormValue("userno")

	if eventid == "" {
		eventid = "1"
	}
	if userno == "" {
		userno = "0"
	}

	fetchedData := fetchDataFromDB(eventid, userno)

	isAjax := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	pageData := PageData{
		EventID: eventid,
		UserNo:  userno,
		Data:    fetchedData,
		IsAjax:  isAjax,
	}

	if isAjax {
		// err := tmpl.ExecuteTemplate(w, "data_table.html", pageData)
		err := tmpl.ExecuteTemplate(w, "data_table", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err := tmpl.ExecuteTemplate(w, "index.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ユーザー名検索を行う新しいハンドラー
func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	queryName := r.FormValue("name") // 検索文字列を取得

	// ユーザー名検索を実行
	foundUsers := searchUsersFromDBByName(queryName)

	// レスポンスヘッダーにContent-Type: application/jsonを設定
	w.Header().Set("Content-Type", "application/json")

	// 検索結果をJSON形式でエンコードして返す
	if err := json.NewEncoder(w).Encode(foundUsers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// データベースからデータを取得するダミー関数 (前回と同じ)
func fetchDataFromDB(eventid, userno string) []Data {
	log.Printf("Searching main data with eventid: %s, userno: %s", eventid, userno)
	var data []Data
	for i := 0; i < 3; i++ {
		data = append(data, Data{
			ID:    i + 1,
			Name:  "Item " + strconv.Itoa(i+1),
			Value: "Event: " + eventid + ", User: " + userno + ", Data: " + strconv.Itoa(i+1),
		})
	}
	return data
}

// ユーザー名で検索するダミー関数
func searchUsersFromDBByName(name string) []User {
	log.Printf("Searching users with name: %s", name)
	// 実際にはここでDBアクセスを行い、部分一致検索を実行
	allUsers := []User{
		{UserNo: "101", UserName: "田中 太郎"},
		{UserNo: "102", UserName: "山田 花子"},
		{UserNo: "103", UserName: "鈴木 一郎"},
		{UserNo: "201", UserName: "佐藤 次郎"},
		{UserNo: "202", UserName: "高橋 三郎"},
		{UserNo: "301", UserName: "田中 次郎"}, // 同姓同名も考慮
	}

	var results []User
	if name == "" {
		// 検索文字列が空の場合は全件返すか、何も返さないか、要件による
		// 今回は全件返す例
		return allUsers
	}

	for _, user := range allUsers {
		// 大文字小文字を区別しない部分一致検索
		if strings.Contains(strings.ToLower(user.UserName), strings.ToLower(name)) {
			results = append(results, user)
		}
	}
	return results
}

/*
```

**`templates/index.html`** (ページ全体を構成するテンプレート)

```html

```

**`templates/data_table.html`** (前回と同じ)

```html

```

---

### 変更点の解説

1.  **Go側の変更 (`main.go`)**:
    *   `User` 構造体を追加し、JSONエンコード時にフィールド名が小文字になるように `json:"userno"` などのタグを付けました。
    *   `searchUsersFromDBByName(name string) []User` 関数を追加しました。これはダミーデータに対して `strings.Contains` を使って部分一致検索を行います。実際にはデータベースの `LIKE` 句などを使用します。
    *   新しいハンドラー `searchUsersHandler(w http.ResponseWriter, r *http.Request)` を追加しました。
        *   `r.FormValue("name")` で検索クエリを取得します。
        *   `searchUsersFromDBByName` を呼び出してユーザーリストを取得します。
        *   `w.Header().Set("Content-Type", "application/json")` でレスポンスがJSONであることをブラウザに伝えます。
        *   `json.NewEncoder(w).Encode(foundUsers)` でユーザーリストをJSON形式でレスポンスボディに書き込みます。
    *   `http.HandleFunc("/search-users", searchUsersHandler)` でこの新しいハンドラーを登録しました。

2.  **HTML/CSSの変更 (`templates/index.html`)**:
    *   ポップアップのタイトルを「ユーザ検索」に変更しました。
    *   `userno` 入力フィールドを削除し、代わりに `searchUserName` というIDのテキスト入力フィールドと `searchUserBtn` というIDの検索ボタンを追加しました。
    *   検索結果を表示するための空の `<ul>` 要素 (`id="searchResultsList"`) と、結果がない場合に表示するメッセージ (`id="noResultsMessage"`) を追加しました。
    *   検索結果から選択された `userno` を使ってメイン画面を更新するための「選択して再表示」ボタン (`id="selectUserAndRefreshBtn"`) を追加しました。このボタンは初期状態では `disabled` にしておき、ユーザーがリストから選択したときに有効化します。
    *   検索結果リスト (`#searchResultsList`) と選択されたアイテム (`.selected`) のためのCSSを追加しました。

3.  **JavaScriptの変更 (`templates/index.html` 内の `<script>`)**:
    *   **ポップアップ表示時**: 検索フィールドや結果リストをクリアし、選択ボタンを無効化する処理を追加しました。`selectedUserNo` 変数もリセットします。
    *   **`searchUserBtn` (検索ボタン) クリック時**:
        *   `searchUserNameInput.value` から検索文字列を取得します。
        *   `fetch(`/search-users?name=${encodeURIComponent(nameQuery)}`)` を使って、新しい `/search-users` エンドポイントにAjaxリクエストを送信します。`encodeURIComponent` で検索文字列をURLエンコードしています。
        *   レスポンスはJSON形式なので、`response.json()` でパースします。
        *   取得した `users` 配列をループし、各ユーザーに対して `<li>` 要素を作成します。
        *   `li.dataset.userno = user.userno;` のように、カスタムデータ属性 (`data-userno`) を使って `userno` をHTML要素に埋め込みます。これにより、後で選択されたときに簡単に `userno` を取得できます。
        *   各 `<li>` にクリックイベントリスナーを追加し、クリックされたらそのアイテムを選択状態にし、`selectedUserNo` 変数に `userno` を格納し、「選択して再表示」ボタンを有効化します。
        *   検索結果がない場合は `noResultsMessage` を表示します。
    *   **`selectUserAndRefreshBtn` (選択して再表示ボタン) クリック時**:
        *   `selectedUserNo` に値があるか確認します。
        *   前回の「再表示」ボタンと同じロジックで、`currentEventId` と `selectedUserNo` を使ってメイン画面のデータ更新Ajaxリクエスト (`/?eventid=...&userno=...`) を送信します。
        *   レスポンスのHTMLフラグメントで `dataDisplayArea` を更新し、`displayedUserNoSpan` も更新します。
        *   ポップアップを閉じます。

---

### 実行方法

1.  `main.go` と `templates` ディレクトリ（その中に `index.html` と `data_table.html`）を作成します。
2.  `go run main.go` でサーバーを起動します。
3.  ブラウザで `http://localhost:8080` にアクセスします。

「ユーザ検索」ボタンを押すとポップアップが開き、ユーザー名を入力して「検索」ボタンを押すと、検索結果がリストで表示されます。リストからユーザーを選択し、「選択して再表示」ボタンを押すと、メイン画面のデータが更新されます。

この構成で、ご要望の「ポップアップ内でユーザー検索を行い、結果を選択してメイン画面を更新する」という複雑なフローを実現できます。
*/
