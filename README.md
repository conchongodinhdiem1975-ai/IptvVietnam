# ⚡ Go IPTV VIP Gateway

Hệ thống Proxy IPTV tối ưu hóa luồng phát chất lượng cao, viết bằng ngôn ngữ **Golang (Native)** và triển khai hoàn toàn miễn phí trên nền tảng **Vercel Serverless**.

## 🚀 Tính năng nổi bật (VIP Edition)
* **Golang Core Engine:** Tốc độ phản hồi cực nhanh, thời gian khởi động nguội gần như bằng 0.
* **Concurrency Dashboard:** Sử dụng Goroutines để quét trạng thái sống/chết của toàn bộ kênh song song trong < 1 giây.
* **Smart In-Memory Cache:** Tự động lưu tạm cấu trúc `.m3u8` trong 10 giây trên RAM, giảm độ trễ khi chuyển kênh về mức tối thiểu.
* **Smart Abort Controller:** Tự động ngắt tiến trình tải luồng từ máy chủ gốc ngay khi bạn bấm chuyển kênh trên Tivi, tiết kiệm băng thông tuyệt đối.

## 📂 Cấu trúc dự án
```text
├── 📂 api
│   └── 📄 index.go       # Bộ não xử lý Proxy & Dashboard
├── 📄 go.mod             # Khai báo cấu hình module Go
├── 📄 vercel.json        # Định tuyến tài nguyên hệ thống
└── 📄 README.md          # Hướng dẫn sử dụng này

