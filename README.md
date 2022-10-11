# annict2anilist

[Annict](https://annict.com) から [AniList](https://anilist.co) にライブラリを同期します。

<img width="976" alt="image" src="https://user-images.githubusercontent.com/7302150/195117351-e58a0831-2fae-44fa-8e0d-42e1567d46bd.png">

同期は次のルールで行われます。

- Annict 側がマスターとなり、「視聴ステータス」「話数」が同期されます。
  - Annict 側では登録されているが、AniList で記録がない場合は作成されます。
  - AniList 側では登録されているが、Annict 側で記録がない場合は何もしません。(Annict のデータを操作することはありません。)
- [kawaiioverflow/arm](https://github.com/kawaiioverflow/arm) を利用して、作品の紐付けを行っています。紐付けができなかった作品データは `untethered.json` に出力されます。

annict2anilist は [ci7lus/imau](https://github.com/ci7lus/imau) の CLI バージョンです。

## 環境変数

以下の環境変数を `.env.example` を参考に `.env` に記述してください。

| 環境変数                                            | Default | Description                                                                                                                                      |
|-------------------------------------------------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| `ANNICT_CLIENT_ID`<br/>`ANNICT_CLIENT_SECRET`   | *必須*    | Annict の OAuth クライアントです。[ここ](https://annict.com/oauth/applications) で発行できます。<br/>リダイレクト URI には `urn:ietf:wg:oauth:2.0:oob` を指定してください。            |
| `ANILIST_CLIENT_ID`<br/>`ANILIST_CLIENT_SECRET` | *必須*    | AniList の OAuth クライアントです。[ここ](https://anilist.co/settings/developer) で発行できます。<br/>リダイレクト URI には `https://anilist.co/api/v2/oauth/pin` を指定してください。 |
| `TOKEN_DIRECTORY`                               | `.`     | トークン情報を格納するディレクトリを指定します。<br/>未指定の場合はカレントディレクトリに格納します。                                                                                            |
| `INTERVAL_MINUTES`                              | `0`     | 指定した分ごとに同期を行います。<br/>未指定の場合は一度同期して終了します。                                                                                                         |
| `DRY_RUN`                                       | `0`     | `1` を指定すると書き込みリクエストを送信しません。デバッグ用です。                                                                                                              |

## Build

```console
$ make build
```

## Run

```console
$ make run
```
