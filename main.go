/*
ご質問ありがとうございます。GoのWebサーバーで、`userno` 入力部分をポップアップ化し、画面遷移なしでデータを更新したいというご要望ですね。

まさにその通りで、このようなケースでは**フロントエンド（JavaScript）でポップアップを実装し、非同期通信（Ajax）を使ってサーバーから新しいデータを取得して画面の一部を更新する**のが最も一般的なアプローチであり、ユーザー体験も向上します。

「操作の途中で気が変わってやり直すときデータ検索が再度発生し無駄な処理が必要になる」という懸念も、ポップアップを閉じれば元の画面状態が維持されるため、この方法で解決できます。

以下に、具体的な実装方法の概要とコード例を示します。

---

### 実装の全体像

1.  **HTML/CSS:**
    *   `userno` 入力用のモーダルダイアログ（ポップアップ）のHTML構造とCSSを作成します。
    *   メイン画面には、ポップアップを開くボタンと、検索結果を表示する領域を用意します。
2.  **JavaScript:**
    *   ポップアップの表示/非表示を制御します。
    *   ポップアップ内の「再表示」ボタンが押されたら、入力された `userno` と現在の `eventid` を使って、サーバーにAjaxリクエストを送信します。
    *   サーバーからのレスポンス（新しいデータ）を受け取り、メイン画面の検索結果表示領域をJavaScriptで更新します。
3.  **Goハンドラー:**
    *   既存のハンドラーを少し修正し、リクエストが通常のページロードなのか、Ajaxリクエストなのかを判別できるようにします。
    *   Ajaxリクエストの場合は、ページ全体ではなく、更新が必要なデータ部分（HTMLフラグメントまたはJSONデータ）のみを返します。

---

### 具体的な実装例

#### 1. Goのハンドラーとテンプレートの準備

まず、Goのハンドラーとテンプレートを以下のように準備します。


v0.0.0  gemini-2.5-flashによるのコード生成(id=2308)
*/

package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv" // usernoを数値として扱う場合

	"crypto/tls"
	"time"
)

// 検索結果のデータ構造 (例)
type Data struct {
	ID    int
	Name  string
	Value string
}

// テンプレートに渡すデータ構造
type PageData struct {
	EventID       string
	UserNo        string // 現在表示されているuserno
	InitialUserNo string // ポップアップのinputフィールドの初期値用
	Data          []Data
	IsAjax        bool // Ajaxリクエストかどうかをテンプレートに伝える (今回はJSで処理するので不要だが、サーバー側でレスポンスを切り替えるために使う)
}

var tmpl *template.Template

func init() {
	// テンプレートファイルのパスは適宜調整してください
	// index.html: ページ全体
	// data_table.html: 検索結果のテーブル部分のみ
	tmpl = template.Must(template.ParseFiles("templates/index.html", "templates/data_table.html"))
}

func main() {
	http.HandleFunc("/", handler)
	// log.Fatal(http.ListenAndServe(":8000", nil))

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

func handler(w http.ResponseWriter, r *http.Request) {
	eventid := r.FormValue("eventid")
	userno := r.FormValue("userno")

	// eventidが指定されていない場合は初期値 (例: "1") を設定
	if eventid == "" {
		eventid = "1"
	}
	// usernoが指定されていない場合は初期値"0"を設定
	if userno == "" {
		userno = "0"
	}

	// ここでデータベースからデータを取得するロジック
	fetchedData := fetchDataFromDB(eventid, userno)

	// リクエストヘッダーでAjaxリクエストかどうかを判別
	isAjax := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	pageData := PageData{
		EventID:       eventid,
		UserNo:        userno, // 現在表示されているuserno
		InitialUserNo: userno, // ポップアップのinputフィールドの初期値
		Data:          fetchedData,
		IsAjax:        isAjax,
	}

	if isAjax {
		// Ajaxリクエストの場合、データ部分のみをレンダリングして返す
		// data_table.html は検索結果のテーブル部分のみを含むテンプレートとする
		err := tmpl.ExecuteTemplate(w, "data_table", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// 通常のリクエストの場合、ページ全体をレンダリング
		err := tmpl.ExecuteTemplate(w, "index.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// データベースからデータを取得するダミー関数
func fetchDataFromDB(eventid, userno string) []Data {
	log.Printf("Searching with eventid: %s, userno: %s", eventid, userno)
	// 実際にはここでDBアクセスを行う
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

/*
### 解説

1.  **Goハンドラー (`main.go`)**:
    *   `eventid` と `userno` をURLパラメータから取得します。
    *   `r.Header.Get("X-Requested-With") == "XMLHttpRequest"` を使って、リクエストがJavaScriptからのAjaxリクエストかどうかを判別します。
    *   **Ajaxリクエストの場合**: `tmpl.ExecuteTemplate(w, "data_table.html", pageData)` を呼び出し、検索結果のテーブル部分のみをレンダリングしたHTMLフラグメントをレスポンスとして返します。
    *   **通常のページロードの場合**: `tmpl.ExecuteTemplate(w, "index.html", pageData)` を呼び出し、ページ全体をレンダリングします。
    *   `fetchDataFromDB` はダミー関数ですが、実際にはここでデータベースアクセスを行います。

2.  **メインテンプレート (`templates/index.html`)**:
    *   `userno` 入力用のモーダルダイアログのHTML (`<div id="userNoModal" class="modal">...</div>`) と、それを開くボタン (`<button id="openUserNoModal">`) を配置します。
    *   検索結果を表示する領域 (`<div id="dataDisplayArea">`) を用意し、初回ロード時には `{{ template "data_table" . }}` で初期データを表示します。
    *   `eventid` はポップアップで変更されないため、`<input type="hidden" id="currentEventId" value="{{ .EventID }}">` のように隠しフィールドで保持し、JavaScriptからアクセスできるようにします。
    *   `displayedUserNo` という `<span>` タグで現在の `userno` を表示し、JavaScriptで更新できるようにします。
    *   **CSS**: モーダルダイアログの表示/非表示やスタイルを定義します。
    *   **JavaScript**:
        *   `DOMContentLoaded` イベントで、DOMが完全にロードされた後に処理を開始します。
        *   `openUserNoModal` ボタンがクリックされたら、モーダル (`userNoModal`) を表示します。この際、`inputUserNo` の初期値として、現在表示されている `userno` をセットします。
        *   モーダルの閉じるボタン (`close-button`) や、モーダルの外側をクリックしたときにモーダルを非表示にします。
        *   `submitUserNo` ボタンがクリックされたら、`inputUserNo` の値と `currentEventId` を取得し、`fetch` APIを使ってサーバーにAjaxリクエストを送信します。
        *   `fetch` リクエストの `headers` に `'X-Requested-With': 'XMLHttpRequest'` を含めることで、サーバー側でAjaxリクエストであることを判別できるようにします。
        *   サーバーからのレスポンス（HTMLフラグメント）を受け取ったら、`dataDisplayArea.innerHTML = htmlFragment;` でDOMを直接更新し、`displayedUserNoSpan.textContent = newUserNo;` でメイン画面の `userno` 表示も更新します。
        *   最後にモーダルを閉じます。

3.  **データテーブルテンプレート (`templates/data_table.html`)**:
    *   検索結果のテーブル部分のみを定義したテンプレートです。GoハンドラーがAjaxリクエスト時にこれをレンダリングして返します。

---

### 実行方法

1.  `main.go` と `templates` ディレクトリ（その中に `index.html` と `data_table.html`）を作成します。
2.  `go run main.go` でサーバーを起動します。
3.  ブラウザで `http://localhost:8080` にアクセスします。

初期データが表示され、「userno入力」ボタンを押すとポップアップが表示されます。ポップアップで `userno` を変更して「再表示」ボタンを押すと、画面全体がリロードされることなく、検索結果のテーブル部分と表示されている `userno` が更新されるはずです。

この方法で、ユーザーの要望通りの機能を実現できるかと思います。
*/
