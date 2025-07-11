# 🚀 localvnc

**Stream your Linux desktop to any device on your local network — instantly, efficiently, and with zero setup!**

![Animated Demo](https://raw.githubusercontent.com/nathfavour/localvnc/main/animated-preview.gif)

---

## ✨ Features

- **Ultra-fast streaming**: Only changed regions are sent, thanks to X11's XDamage extension.
- **No VNC client/server needed**: Just run the binary and open your browser!
- **Single binary**: Everything (server + HTML5 client) is embedded.
- **QR code access**: Scan and connect from your phone or tablet.
- **Live HTML5 client**: View your desktop in any browser.
- **Adaptive quality**: Smart region merging and JPEG compression.
- **Works everywhere**: Laptops, tablets, phones — anything with a browser.

---

## 🎬 How It Works

1. **Start the server:**
   ```bash
   ./localvnc --port 3456
   ```
2. **Scan the QR code or open the URL on any device:**
   ```
   http://<your-ip>:3456/client
   ```
3. **Enjoy live, animated streaming of your desktop!**

---

## 🖥️ Animated Preview

<p align="center">
  <img src="https://raw.githubusercontent.com/nathfavour/localvnc/main/animated-preview.gif" alt="localvnc animated demo" width="600"/>
</p>

---

## ⚡️ Tech Highlights

- **Go + X11**: Uses X11's XDamage for efficient change detection.
- **Delta frames**: Only changed regions are sent, saving bandwidth.
- **Embedded HTML5 client**: No need for external files.
- **Smart region merging**: Fewer, larger updates for smooth animation.
- **Frame buffering**: Always serves the latest frame for instant response.

---

## 🚦 Usage

```bash
# Run on any Linux desktop with X11
./localvnc --port 3456

# Open http://<your-ip>:3456/client in your browser
```

---

## 🛠️ Build

```bash
go build -o localvnc ./cmd/localvnc
```

---

## 💡 Tips

- For best performance, use a wired connection or fast WiFi.
- Try on multiple devices — it's fun!
- Adjust JPEG quality and frame rate for your network.

---

## 🤝 Contributing

Pull requests and issues welcome!  
Let's make localvnc the fastest, easiest desktop streamer for Linux.

---

## 📜 License

MIT

---

<p align="center">
  <b>localvnc</b> — <i>your desktop, everywhere, instantly.</i>
</p>
