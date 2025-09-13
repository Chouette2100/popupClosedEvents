// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
/*
ClosedEventsHandler()等において、ユーザー名あるいはイベント名から検索して該当するイベント一覧を表示するための手法の検討

Gemini-2.5-flashによるのコード生成(id=2321)

はい、承知いたしました。ポップアップが複数あるケース、特に「イベント検索」と「ユーザー検索」の2つのモーグダルダイアログを連携させる例を作成します。

このケースでは、単に処理を二つ並べるだけでなく、以下の点に工夫が必要です。

1.  **グローバルな状態管理**: 現在選択されている `eventid` と `userno` をJavaScriptの変数として管理し、どちらのモーダルからでも参照・更新できるようにします。
2.  **モーダルの排他制御**: 複数のモーダルが同時に開かないように、一方を開くときに他方を閉じる処理を入れます。
3.  **メイン画面更新の共通化**: どちらのモーダルで値が選択されても、最終的にメイン画面のデータを更新する処理は共通化します。

---

### 実装の全体像

1.  **Goハンドラーの追加**:
  - イベント名検索リクエストを受け付け、`eventno` と `eventname` のリストをJSON形式で返す新しいエンドポイント（例: `/search-events`）を追加します。
  - 既存の `/search-users` ハンドラーはそのまま使用します。

2.  **HTML/CSSの変更**:
  - 「イベント検索」ボタンと、それに対応する新しいモーダルダイアログ (`id="eventSearchModal"`) を追加します。
  - 既存の「ユーザー検索」モーダル (`id="userSearchModal"`) はそのまま使用します。
  - メイン画面に表示される `Event ID` と `User No` の表示箇所を更新します。

3.  **JavaScriptの変更**:
  - `currentEventId` と `currentUserNo` をJavaScriptの変数として管理します。
  - イベント検索モーダルとユーザー検索モーダルの開閉ロジックをそれぞれ実装します。
  - 各モーダルを開く際に、もう一方のモーダルがもし開いていたら閉じる処理を追加します。
  - イベント検索モーダル内でイベント名検索を行い、結果リストから選択された `eventno` を `currentEventId` に設定します。
  - ユーザー検索モーダル内でユーザー名検索を行い、結果リストから選択された `userno` を `currentUserNo` に設定します。
  - `currentEventId` または `currentUserNo` が更新されたら、共通の `refreshMainData()` 関数を呼び出してメイン画面のデータを更新します。

---

### 具体的な実装例

#### 1. Goのハンドラーとテンプレートの準備

**`main.go`**

```go

v0.0.0  ユーザー番号によりイベントを絞り込み一覧を表示する。gemini-2.5-flashでコードを生成する(id=2308)
v0.0.1  .gitignoreを追加する
v0.1.0  ユーザー番号による検索から、ユーザー名による絞り込みに変更する（id=2311）
v0.1.2  ユーザー名とイベント名の絞り込みがあり、ダイアログが複数必要なケースを作成する（id=2321）

*/
package main

import (
	"crypto/tls"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
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

// イベント検索結果のデータ構造 (新規追加)
type Event struct {
	EventNo   string `json:"eventno"`
	EventName string `json:"eventname"`
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
	http.HandleFunc("/search-users", searchUsersHandler)
	http.HandleFunc("/search-events", searchEventsHandler) // 新しいハンドラーを追加
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
		eventid = "1" // 初期値
	}
	if userno == "" {
		userno = "0" // 初期値
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

// ユーザー名検索を行うハンドラー (前回と同じ)
func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	queryName := r.FormValue("name")

	foundUsers := searchUsersFromDBByName(queryName)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(foundUsers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// イベント名検索を行う新しいハンドラー
func searchEventsHandler(w http.ResponseWriter, r *http.Request) {
	queryName := r.FormValue("name")

	foundEvents := searchEventsFromDBByName(queryName)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(foundEvents); err != nil {
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

// ユーザー名で検索するダミー関数 (前回と同じ)
func searchUsersFromDBByName(name string) []User {
	log.Printf("Searching users with name: %s", name)
	allUsers := []User{
		{UserNo: "101", UserName: "田中 太郎"},
		{UserNo: "102", UserName: "山田 花子"},
		{UserNo: "103", UserName: "鈴木 一郎"},
		{UserNo: "201", UserName: "佐藤 次郎"},
		{UserNo: "202", UserName: "高橋 三郎"},
		{UserNo: "301", UserName: "田中 次郎"},
	}

	var results []User
	if name == "" {
		return allUsers
	}

	for _, user := range allUsers {
		if strings.Contains(strings.ToLower(user.UserName), strings.ToLower(name)) {
			results = append(results, user)
		}
	}
	return results
}

// イベント名で検索するダミー関数 (新規追加)
func searchEventsFromDBByName(name string) []Event {
	log.Printf("Searching events with name: %s", name)
	allEvents := []Event{
		{EventNo: "1", EventName: "新入社員研修"},
		{EventNo: "2", EventName: "Go言語勉強会"},
		{EventNo: "3", EventName: "WebAssembly入門"},
		{EventNo: "4", EventName: "社内ハッカソン"},
		{EventNo: "5", EventName: "Go開発者会議"},
		{EventNo: "10", EventName: "Go開発者会議2024"},
	}

	var results []Event
	if name == "" {
		return allEvents
	}

	for _, event := range allEvents {
		if strings.Contains(strings.ToLower(event.EventName), strings.ToLower(name)) {
			results = append(results, event)
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

#### Go側の変更 (`main.go`)

*   `Event` 構造体を追加しました。
*   `searchEventsFromDBByName(name string) []Event` ダミー関数を追加しました。これはイベント名で部分一致検索を行います。
*   `searchEventsHandler(w http.ResponseWriter, r *http.Request)` を追加し、`/search-events` エンドポイントでイベント検索結果をJSONで返すようにしました。
*   `main` 関数で `http.HandleFunc("/search-events", searchEventsHandler)` を追加し、新しいハンドラーを登録しました。

#### HTML/CSSの変更 (`templates/index.html`)

*   **イベント検索ボタン**: `<button id="openEventSearchModal">イベント検索</button>` を追加しました。
*   **イベント検索モーダル**:
    *   `id="eventSearchModal"` を持つ新しいモーダルダイアログのHTML構造を追加しました。
    *   内部にはイベント名入力用の `<input type="text" id="searchEventName">`、検索ボタン `<button id="searchEventBtn">`、検索結果表示用の `<ul id="searchEventsList" class="searchResultsList">`、そして「選択して再表示」ボタン `<button id="selectEventAndRefreshBtn">` があります。
*   **メイン画面の表示**: `Event ID` と `User No` の表示を `<strong>` タグ内の `id="displayedEventId"` と `id="displayedUserNo"` に変更し、JavaScriptから更新できるようにしました。
*   **初期値の保持**: `eventid` と `userno` の初期値をJavaScriptで取得するために、`<input type="hidden" id="initialEventId" value="{{ .EventID }}">` と `<input type="hidden" id="initialUserNo" value="{{ .UserNo }}">` を追加しました。
*   **CSSの共通化**: 検索結果リストのスタイルを `.searchResultsList` クラスとして定義し、両方のモーダルで利用するようにしました。ユーザー検索リストのIDも `searchUsersList` に変更し、重複を避けました。

#### JavaScriptの変更 (`templates/index.html` 内の `<script>`)

1.  **グローバル状態変数**:
    *   `let currentEventId` と `let currentUserNo` を定義し、隠しフィールドから初期値を取得するようにしました。これらの変数が、現在メイン画面に表示されているデータに対応するIDを保持します。
2.  **DOM要素の取得**:
    *   新しいイベント検索モーダルに関連するすべてのDOM要素を取得しました。
3.  **共通関数 `closeModal(modalElement)`**:
    *   指定されたモーダル要素を非表示にするシンプルな関数です。
4.  **共通関数 `refreshMainData()`**:
    *   `currentEventId` と `currentUserNo` の現在の値を使って、メイン画面のデータ (`/?eventid=...&userno=...`) をAjaxで再取得し、`dataDisplayArea` を更新します。
    *   `displayedEventIdSpan` と `displayedUserNoSpan` の表示も更新します。
5.  **イベント検索モーダル処理**:
    *   `openEventModalBtn.onclick`: イベント検索モーダルを開く際に、**まず `userModal` を閉じる**処理を追加しました。その後、イベント検索モーダルを初期化して表示します。
    *   `searchEventBtn.onclick`: 入力されたイベント名で `/search-events` エンドポイントにAjaxリクエストを送信し、結果を `searchEventsList` に動的に表示します。
    *   リストの `<li>` 要素がクリックされたら、`selectedEventNo` にそのイベントのIDを格納し、「選択して再表示」ボタンを有効化します。
    *   `selectEventAndRefreshBtn.onclick`: `selectedEventNo` を `currentEventId` に設定し、**`refreshMainData()` を呼び出してメイン画面を更新**します。その後、モーダルを閉じます。
6.  **ユーザー検索モーダル処理**:
    *   `openUserModalBtn.onclick`: ユーザー検索モーダルを開く際に、**まず `eventModal` を閉じる**処理を追加しました。その後、ユーザー検索モーダルを初期化して表示します。
    *   `searchUserBtn.onclick`: ユーザー名検索のロジックは前回と同じですが、検索結果リストのIDを `searchUsersList` に変更しました。
    *   `selectUserAndRefreshBtn.onclick`: `selectedUserNo` を `currentUserNo` に設定し、**`refreshMainData()` を呼び出してメイン画面を更新**します。その後、モーダルを閉じます。
7.  **`window.onclick`**: モーダルの外側をクリックしたときに閉じる処理も、両方のモーダルに対応するように修正しました。

---

### 実行方法

1.  `main.go` と `templates` ディレクトリ（その中に `index.html` と `data_table.html`）を作成します。
2.  `go run main.go` でサーバーを起動します。
3.  ブラウザで `http://localhost:8080` にアクセスします。

「イベント検索」ボタンを押すとイベント検索モーダルが開き、イベント名で検索・選択できます。
「ユーザ検索」ボタンを押すとユーザー検索モーダルが開き、ユーザー名で検索・選択できます。

どちらかのモーダルで選択を行い「選択して再表示」ボタンを押すと、メイン画面のデータが、その時点で確定している `Event ID` と `User No` の両方を使って更新されます。

この構成により、複数のモーダルダイアログが連携し、メイン画面の状態を更新する、より複雑なアプリケーションフローを実現できます。JavaScriptで `currentEventId` と `currentUserNo` を一元的に管理し、`refreshMainData()` 関数でメイン画面の更新処理を共通化している点がポイントです。
*/
