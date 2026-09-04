# quicknews keymap

実測ベース(tui/feedlist.go, tui/articlelist.go, tui/summaryview.go, tui/update_handlers.go)。変更時はこの表を更新する。

## 共通 (update_handlers.go)
| キー | 動作 |
|---|---|
| ctrl+c / q | 終了 |

## feed list view (feedlist.go)
| キー | 動作 |
|---|---|
| enter | feed 選択 → `selectFeedMsg` で article list へ |
| r | feed 再読込 |
| d | feed 削除(確認ダイアログ、bookmark feed は対象外) |

## article list view (articlelist.go)
| キー | 動作 |
|---|---|
| enter | 記事選択 → `selectArticleMsg` で summary view へ |
| b | feed list へ戻る(`backToFeedListMsg`) |
| r | 記事再読込 |
| a | 未読のみ / 全件 表示切替 |
| o | ブラウザで開く |
| R | 既読状態トグル |
| A | 現在の feed の全記事を既読にする |

## summary view (summaryview.go)
| キー | 動作 |
|---|---|
| b | article list へ戻る(`backToArticleListMsg`) |
| j / k / down / up | スクロール |
| g / G | 先頭 / 末尾へ |
| o | ブラウザで開く |
| B | ブックマーク登録 |
| r | 音声再生(確認ダイアログあり。bookmark feed 以外かつ summary がある場合のみ) |
| d | 記事と要約を削除(確認ダイアログ) |

## 遷移フロー
- `selectFeedMsg`: feed → article list
- `selectArticleMsg`: article → summary view(本文取得 cmd を発行)
- `fetchedArticleContentMsg`: summary view に本文をセット
- `backToFeedListMsg` / `backToArticleListMsg`: 各一覧へ戻る
- Update の分岐順: グローバルキー → window size → error → view 遷移 msg → 現在 view へ委譲
