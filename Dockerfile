# build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . . 
# CGO_ENABLED=0:關閉 CGO,讓 Go 產生純靜態連結的執行檔,不依賴系統的 C 函式庫(如 glibc)。
# 這是能用 scratch 這種空白基底映像的關鍵前提

# GOOS=linux:確保跨平台編譯時目標是 Linux(如果你在 M 系列 Mac 上開發,這行能避免編出 macOS 執行檔

# -a -installsuffix cgo:強制重新編譯所有套件(而不用快取),installsuffix cgo 是舊版避免 CGO/non-CGO 套件快取衝突的寫法
#(在 CGO_ENABLED=0 下其實已經比較少必要,但很多範例仍保留這個慣例寫法)

# -o api cmd/api/*.go:把 cmd/api 目錄下的 main package 編譯成一個叫 api 的執行檔
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/*.go

# run stage
# 這是 Docker 裡「空白」的基底映像,裡面什麼都沒有——沒有 shell、沒有套件管理器、沒有作業系統檔案。
# 這也是整個 Dockerfile 的精華所在:最終映像檔可以小到只有幾 MB(因為裡面只有你的執行檔)
FROM scratch
WORKDIR /app
# copy CA certificates from the build stage
# 因為 scratch 什麼都沒有,連 CA 根憑證都沒有,如果你的程式要對外發 HTTPS 請求(例如呼叫 SendGrid API),沒有這個檔案會驗證憑證失敗。所以要從 builder 階段把憑證檔案複製過來
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# 只複製編譯好的執行檔,不帶原始碼、不帶 Go 工具鏈,大幅縮小最終映像檔體積、減少攻擊面。
COPY --from=builder /app/api .
# 純粹是文件性質的宣告,告訴使用者/工具這個容器預期監聽 8080 port(不會實際影響 runtime 行為)
EXPOSE 8080
# 容器啟動時執行的指令
CMD ["./api"]