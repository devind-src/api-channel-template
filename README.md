# api-channel-template

Template untuk interface **inbound** baru: `api-channel-xxx` (contoh: `api-channel-tokopedia`, `api-channel-mobile`).
Channel eksternal memanggil app ini lewat HTTP. App meneruskan transaksi ke synccore, lalu menjawab request HTTP yang masih terbuka setelah synccore merespons.

## Aturan

- Dependensi internal **hanya** `github.com/devind-src/syncnet-sdk`. Ini dijaga oleh `TestDependsOnlyOnSDK`, jadi `go test ./...` akan gagal kalau ada import ke syncnet-core atau app lain.
- Semua logika khusus channel ada di `internal/channel`. HTTP server, koneksi ke synccore, command port, logging, trace, dan lifecycle proses sudah disediakan oleh sdk.
- Versi sdk dikunci di `go.mod`. Upgrade dilakukan dengan sengaja: `go get github.com/devind-src/syncnet-sdk@vX.Y.Z`.

## Membuat channel baru

1. Buat repo dari template ini, misalnya `api-channel-mobile`.
2. Ganti semua nama `template`:
   ```sh
   go mod edit -module github.com/devind-src/api-channel-mobile
   git mv cmd/api-channel-template cmd/api-channel-mobile
   grep -rl 'api-channel-template' . | xargs sed -i 's/api-channel-template/api-channel-mobile/g'
   ```
3. Isi `cmd/api-channel-mobile/appsettings.json`:
   - `AppName` harus sama dengan `app_name` di `sw_app`.
   - Isi juga kredensial channel.
4. Daftarkan app, node, dan koneksi HTTP server-nya di database (`sw_app`, `sw_nodes`, `sw_connections`) lewat dashboard. Port HTTP diatur di sana, bukan di kode.
5. Ubah file sesuai API yang disepakati dengan channel (semua bagian yang perlu diubah ditandai `TODO`):

   | File | Isi |
   |---|---|
   | `config.go` | section `Channel` di appsettings.json |
   | `models.go` | struktur request/response API untuk channel |
   | `auth.go` | skema autentikasi (contoh: HMAC-SHA512 dengan cek timestamp) |
   | `routes.go` | path URL → MsgType/TranType |
   | `tointernal.go` | request channel → `message.Request` |
   | `toremote.go` | `message.Response` → response channel |
   | `respcode.go` | error yang dijawab langsung oleh app (auth, URL, format, duplikat) |
   | `app.go` | alurnya: autentikasi → routing → korelasi → reply (jarang perlu diubah) |

6. Kalau channel butuh fee atau routing ke supplier, pakai `syncnet-sdk/billerrouting` (contohnya ada di `api-channel`, `applyFeesAndRouting`).
7. Lengkapi test di `channel_test.go` dengan contoh request asli dari channel.

## Struktur

```
cmd/api-channel-xxx/       main.go (hanya app.Run) + appsettings.json
internal/channel/          seluruh logika channel (satu package)
go.mod                     require github.com/devind-src/syncnet-sdk vX.Y.Z
```

## Development lokal

```sh
go env -w GOPRIVATE=github.com/devind-src/*
go test ./...
go run ./cmd/api-channel-xxx -config cmd/api-channel-xxx/appsettings.json
```

Untuk mengubah sdk dan app sekaligus, pakai `go.work` di folder induk (`use ./syncnet-sdk ./api-channel-xxx`). File itu hanya untuk dev lokal dan tidak di-commit ke repo app.
