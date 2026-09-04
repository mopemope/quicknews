# quicknews Testing Conventions

## パッケージ種別ごとのパターン

### repository / 永続化 (`models/`, `database/`)
- `enttest.Open` + in-memory SQLite で実 DB を立てる。fake で DB を置き換えない。
- `database/` のみ `sqlmock` を使う(DSN 組み立て・接続文字列の確認用)。
- eager loading(`WithFeed()` / `WithSummary()`)の有無もここで検証する。

### command 層 (`cmd/`, `cmd/fetch/`)
- repository / storage は手書き fake に差し替える。mock generator は使わない。
- fake には `var _ Interface = (*fake)(nil)` を付けてコンパイル時に interface 準拠を保証する。
- 分岐・日付処理・正規化などの pure function は table-driven(`t.Run(tt.name, ...)`)で書く。
- 外部依存(merge 音声、audio 保存)は関数フィールド(`mergeAudio` / `saveAudioData`)として注入可能に保つ。

### parser / 変換 (`gemini/` の parse 系, `mcpserver/` の normalize 系, `org/`)
- table-driven。入出力の対応が一覧できる形にする。

### 外部連携 (`gemini/` 実 API, `tts/`, `storage/`)
- 認証情報や endpoint が未設定のときは `t.Skip` する。この前提を壊さない(CI は credential 無しで動く)。
- ロジック本体は fake / interface 越しにテストする。

## 共通ルール
- assertion は testify(`require` / `assert`)。
- エラーメッセージの検証には `require.ErrorContains` を使う。
- test file は対象と同じ package 配下の `*_test.go` に置く(独立 test package にしない)。
- 新規テストは上記のどのパターンに該当するかを判断してから書く。混在させない。

## 検証コマンドの選び方
- 変更パッケージの test をまず回す。影響範囲ごとの最小セットは [verification-matrix.md](verification-matrix.md) を参照。
- 並行処理に触れたら `make test-race`。
- 最終確認は `make test-all`。
