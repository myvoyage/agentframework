# 平台依赖和功能支持

本文档详细说明了AgentFramework在不同操作系统平台上的功能支持和依赖要求。

## 目录

- [Windows](#windows)
- [macOS](#macos)
- [Linux](#linux)
- [功能支持矩阵](#功能支持矩阵)
- [安装依赖指南](#安装依赖指南)

---

## Windows

### 剪贴板模块

**依赖要求**：
- PowerShell 5.1+ (Windows 7 SP1+ 内置)

**功能**：
- ✅ 文本读写
- ✅ 图片读写
- ✅ 清空剪贴板
- ✅ 监控剪贴板变化
- ✅ 历史记录

**说明**：
Windows剪贴板模块使用PowerShell的`Get-Clipboard`和`Set-Clipboard` cmdlet，无需额外安装依赖。

### 通知模块

**依赖要求**：
- Windows 10 Build 1607+ (支持Toast通知)
- Windows PowerShell 5.1+

**功能**：
- ✅ 基本通知（标题、正文）
- ✅ 通知图标
- ✅ 通知图片
- ✅ 提示音
- ✅ 优先级
- ✅ 分组
- ✅ 操作按钮
- ✅ 进度条
- ✅ 持久化

**说明**：
Windows通知模块使用Windows Runtime API通过PowerShell脚本调用，提供完整的Toast通知功能。

### 语音合成 (TTS)

**依赖要求**：
- Windows SAPI (Speech API) 内置于Windows
- 可选：FFmpeg（用于音频格式转换）

**功能**：
- ✅ 语音合成
- ✅ 多种语音
- ✅ 语速调整
- ✅ 音调调整
- ✅ 音量调整
- ✅ 输出格式选择 (WAV, MP3, OGG)
- ✅ 质量选择

**支持的语音**：
- Microsoft David (男，美国英语)
- Microsoft Zira (女，美国英语)
- Microsoft Huihui (女，简体中文)
- 其他系统已安装的语音包

### 语音识别 (STT)

**依赖要求**：
- Windows 10+ (内置语音识别)
- FFmpeg（推荐，用于音频处理）
- 可选：Whisper（可选的高级识别引擎）

**功能**：
- ✅ 语音转文字
- ✅ 录音功能
- ✅ 语言检测
- ✅ 标点符号
- ✅ 时间戳
- ✅ 多种音频格式支持

### 工作流调度

**依赖要求**：
- 无额外依赖

**功能**：
- ✅ Cron表达式解析
- ✅ 定时任务调度
- ✅ 静态任务
- ✅ AI任务
- ✅ 下次运行时间计算

### CAN总线

**依赖要求**：
- Linux: SocketCAN支持 (内核配置CONFIG_CAN)
- 可选: `can-utils` 工具包 (用于测试)
- 安装命令：
  ```bash
  # Debian/Ubuntu
  sudo apt-get install can-utils

  # Fedora/RHEL
  sudo dnf install can-utils

  # Arch Linux
  sudo pacman -S can-utils
  ```

**功能**：
- ⚠️ 框架实现 (完整支持需要平台特定驱动)
- ✅ 标准帧 (11位ID)
- ✅ 扩展帧 (29位ID)
- ✅ 消息收发
- ✅ 过滤器配置
- ✅ 批量发送
- ✅ 错误检测

**说明**：
CAN总线模块提供跨平台框架实现。Linux完整支持需要使用SocketCAN，这需要Linux内核支持。
在Windows和macOS上，模块使用模拟实现便于开发和测试。

### GPIO控制

**依赖要求**：
- Linux: sysfs GPIO接口 或 libgpiod
- 需要 root 权限或 gpio 组权限
- 安装命令：
  ```bash
  # Debian/Ubuntu
  sudo apt-get install gpiod

  # Fedora/RHEL
  sudo sudo dnf install libgpiod

  # Arch Linux
  sudo pacman -S libgpiod
  ```

**功能**：
- ⚠️ 框架实现 (完整支持需要平台特定驱动)
- ✅ 引脚数字输入/输出
- ✅ 引脚方向配置
- ✅ 上下拉电阻配置
- ✅ 边沿检测 (rising/falling/both)
- ✅ PWM输出 (硬件支持依赖)
- ✅ 批量操作

**说明**：
GPIO模块提供跨平台框架实现。Linux完整支持需要sysfs GPIO或libgpiod。
在Windows和macOS上，模块使用模拟实现便于开发和测试。
实际硬件操作通常用于树莓派、嵌入式Linux设备等。

---

## macOS

### 剪贴板模块

**依赖要求**：
- `pbcopy` 和 `pbpaste` 命令 (macOS内置)
- `osascript` (macOS内置)

**功能**：
- ✅ 文本读写
- ✅ 图片读写
- ✅ 清空剪贴板
- ✅ 监控剪贴板变化
- ✅ 历史记录

**说明**：
macOS剪贴板模块使用系统自带的`pbcopy`、`pbpaste`和`osascript`工具。

### 通知模块

**依赖要求**：
- macOS 10.9+ (支持通知中心)
- `osascript` (macOS内置)

**功能**：
- ✅ 基本通知（标题、正文）
- ✅ 通知图标
- ✅ 提示音
- ⚠️ 优先级 (部分支持)
- ✅ 分组
- ⚠️ 操作按钮 (部分支持)

**说明**：
macOS通知模块使用AppleScript通过`osascript`调用，功能相比Windows稍有限制。

### 语音合成 (TTS)

**依赖要求**：
- `say` 命令 (macOS内置)
- 可选：FFmpeg（用于音频格式转换）

**功能**：
- ✅ 语音合成
- ✅ 多种语音
- ✅ 语速调整
- ⚠️ 音调调整 (有限支持)
- ✅ 音量调整
- ✅ 输出格式选择 (WAV, MP3, OGG)
- ✅ 质量选择

**支持的语音**：
系统语音可通过`say -v ?`命令查看，常用包括：
- Alex (美国英语，男)
- Samantha (美国英语，女)
- Ting-Ting (简体中文，女)
- Mei-Jia (繁体中文，女)

### 语音识别 (STT)

**依赖要求**：
- macOS 10.15+ (内置语音识别)
- FFmpeg（推荐，用于音频处理）
- 可选：Whisper（可选的高级识别引擎）

**功能**：
- ✅ 语音转文字
- ✅ 录音功能
- ✅ 语言检测
- ⚠️ 标点符号 (有限支持)
- ⚠️ 时间戳 (有限支持)
- ✅ 多种音频格式支持

**说明**：
macOS的语音识别使用内置的Dictation功能，需要用户授权麦克风权限。

### 工作流调度

**依赖要求**：
- 无额外依赖

**功能**：
- ✅ Cron表达式解析
- ✅ 定时任务调度
- ✅ 静态任务
- ✅ AI任务
- ✅ 下次运行时间计算

---

## Linux

### 剪贴板模块

**依赖要求**：
- `xclip` 或 `xsel` (至少一个)
- 安装命令：
  ```bash
  # Debian/Ubuntu
  sudo apt-get install xclip
  # 或
  sudo apt-get install xsel

  # Fedora/RHEL
  sudo dnf install xclip
  # 或
  sudo dnf install xsel

  # Arch Linux
  sudo pacman -S xclip
  # 或
  sudo pacman -S xsel
  ```

**功能**：
- ✅ 文本读写
- ✅ 图片读写 (需要图形环境)
- ✅ 清空剪贴板
- ✅ 监控剪贴板变化 (部分支持)
- ✅ 历史记录

**说明**：
Linux剪贴板功能依赖X11或Wayland的剪贴板工具，需要在图形环境中运行。

### 通知模块

**依赖要求**：
- `notify-send` (libnotify包)
- 安装命令：
  ```bash
  # Debian/Ubuntu
  sudo apt-get install libnotify-bin

  # Fedora/RHEL
  sudo dnf install libnotify

  # Arch Linux
  sudo pacman -S libnotify
  ```

**功能**：
- ✅ 基本通知（标题、正文）
- ✅ 通知图标
- ⚠️ 通知图片 (部分支持)
- ⚠️ 提示音 (部分支持)
- ⚠️ 优先级 (部分支持)
- ⚠️ 分组 (部分支持)
- ❌ 操作按钮 (不支持)

**说明**：
Linux通知功能依赖桌面环境的通知守护进程（如GNOME、KDE、XFCE等）。

### 语音合成 (TTS)

**依赖要求**：
- `espeak` 或 `festival` (至少一个)
- 推荐安装：
  ```bash
  # Debian/Ubuntu
  sudo apt-get install espeak espeak-ng
  # 或
  sudo apt-get install festival festvox-us-all

  # Fedora/RHEL
  sudo dnf install espeak
  # 或
  sudo dnf install festival

  # Arch Linux
  sudo pacman -S espeak-ng
  # 或
  sudo pacman -S festival festival-us
  ```
- FFmpeg（推荐，用于音频格式转换）

**功能**：
- ⚠️ 语音合成 (质量一般)
- ⚠️ 多种语音 (有限)
- ✅ 语速调整
- ❌ 音调调整 (不支持)
- ✅ 音量调整
- ✅ 输出格式选择 (WAV, MP3, OGG)
- ✅ 质量选择

**说明**：
Linux的TTS功能使用espeak或festival，合成质量相对较低。建议考虑使用第三方TTS服务API。

### 语音识别 (STT)

**依赖要求**：
- `ffmpeg` (必需，用于音频处理)
- `pocketsphinx` 或 `whisper` (至少一个)
- 安装命令：
  ```bash
  # 安装FFmpeg
  sudo apt-get install ffmpeg

  # 安装Pocketsphinx (轻量级)
  sudo apt-get install pocketsphinx pocketsphinx-en-us

  # 或安装Whisper (高质量，需要GPU)
  pip install openai-whisper
  ```
- 可选：`sox`（用于高级音频处理）

**功能**：
- ✅ 语音转文字 (Whisper质量高，Pocketsphinx一般)
- ✅ 录音功能 (需要音频设备)
- ✅ 语言检测 (Whisper支持)
- ⚠️ 标点符号 (Whisper支持)
- ✅ 时间戳 (Whisper支持)
- ✅ 多种音频格式支持

**说明**：
Linux的STT功能需要额外的模型下载和配置。Pocketsphinx相对轻量，Whisper识别质量更高但需要更多资源。

### 工作流调度

**依赖要求**：
- 无额外依赖

**功能**：
- ✅ Cron表达式解析
- ✅ 定时任务调度
- ✅ 静态任务
- ✅ AI任务
- ✅ 下次运行时间计算

---

## 功能支持矩阵

| 功能 | Windows | macOS | Linux | 备注 |
|------|---------|-------|-------|------|
| **剪贴板** |
| 文本读写 | ✅ | ✅ | ✅ | Linux需要xclip/xsel |
| 图片读写 | ✅ | ✅ | ⚠️ | Linux需要图形环境 |
| 历史记录 | ✅ | ✅ | ✅ | |
| **通知** |
| 基本通知 | ✅ | ✅ | ✅ | Linux需要notify-send |
| 优先级 | ✅ | ⚠️ | ⚠️ | 依赖系统支持 |
| 分组 | ✅ | ✅ | ⚠️ | 依赖系统支持 |
| 操作按钮 | ✅ | ⚠️ | ❌ | |
| **TTS** |
| 语音合成 | ✅ | ✅ | ⚠️ | Linux质量一般 |
| 多语音 | ✅ | ✅ | ⚠️ | Linux有限 |
| 音调调整 | ✅ | ⚠️ | ❌ | |
| 输出格式 | ✅ | ✅ | ✅ | 需要FFmpeg |
| **STT** |
| 语音识别 | ✅ | ✅ | ⚠️ | Linux需要模型 |
| 录音 | ✅ | ✅ | ⚠️ | 需要音频设备 |
| 语言检测 | ⚠️ | ✅ | ✅ | Windows部分支持 |
| 时间戳 | ⚠️ | ⚠️ | ✅ | 需要模型支持 |
| **调度** |
| Cron解析 | ✅ | ✅ | ✅ | 无平台差异 |
| 定时任务 | ✅ | ✅ | ✅ | |
| **硬件** |
| CAN总线 | ✅ | ❌ | ⚠️ | Linux需SocketCAN |
| GPIO | ✅ | ⚠️ | ⚠️ | Linux需sysfs/libgpiod |

图例：
- ✅ 完全支持
- ⚠️ 部分支持或有限制/模拟实现
- ❌ 不支持

---

## 安装依赖指南

### Windows

Windows系统通常无需额外安装依赖，系统自带所需功能。

**可选安装**：
```powershell
# 安装FFmpeg（用于音频格式转换）
# 下载：https://ffmpeg.org/download.html
# 解压后将bin目录添加到PATH环境变量
```

### macOS

**安装基础工具**：
```bash
# macOS自带所需工具，无需额外安装
# pbcopy, pbpaste, osascript, say 均为系统内置
```

**可选安装**：
```bash
# 使用Homebrew安装FFmpeg
brew install ffmpeg

# 安装其他可选工具
brew install sox
```

### Linux

**Ubuntu/Debian**：
```bash
# 剪贴板支持
sudo apt-get install xclip

# 通知支持
sudo apt-get install libnotify-bin

# TTS支持
sudo apt-get install espeak espeak-ng

# STT支持
sudo apt-get install ffmpeg pocketsphinx

# 可选：高质量TTS/STT
pip install openai-whisper
```

**Fedora/RHEL**：
```bash
# 剪贴板支持
sudo dnf install xclip

# 通知支持
sudo dnf install libnotify

# TTS支持
sudo dnf install espeak

# STT支持
sudo dnf install ffmpeg pocketsphinx
```

**Arch Linux**：
```bash
# 剪贴板支持
sudo pacman -S xclip

# 通知支持
sudo pacman -S libnotify

# TTS支持
sudo pacman -S espeak-ng

# STT支持
sudo pacman -S ffmpeg pocketsphinx
```

### 检查依赖

**Windows**：
```powershell
# 检查PowerShell版本
$PSVersionTable.PSVersion

# 检查FFmpeg
ffmpeg -version
```

**macOS**：
```bash
# 检查剪贴板工具
which pbcopy pbpaste osascript

# 检查TTS工具
which say

# 检查FFmpeg
ffmpeg -version
```

**Linux**：
```bash
# 检查剪贴板工具
which xclip xsel

# 检查通知工具
which notify-send

# 检查TTS工具
which espeak festival

# 检查STT工具
which ffmpeg pocketsphinx
```

---

## 故障排除

### Windows

**问题**：剪贴板操作失败
- **解决**：确保PowerShell可以运行，检查执行策略：
  ```powershell
  Get-ExecutionPolicy
  # 如果受限，设置为RemoteSigned：
  Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```

**问题**：通知不显示
- **解决**：检查Windows通知设置：
  1. 打开"设置" > "系统" > "通知和操作"
  2. 确保"获取来自应用和其他发送者的通知"已开启
  3. 允许AgentFramework显示通知

### macOS

**问题**：TTS语音不自然
- **解决**：在"系统偏好设置" > "辅助功能" > "语音"中设置默认语音

**问题**：通知不显示
- **解决**：在"系统偏好设置" > "通知"中允许AgentFramework显示通知

**问题**：STT无法使用
- **解决**：在"系统偏好设置" > "键盘" > "听写"中启用听写功能，并授权麦克风权限

### Linux

**问题**：剪贴板工具无法使用
- **解决**：
  ```bash
  # 检查是否在图形环境中
  echo $DISPLAY

  # 安装xclip或xsel
  sudo apt-get install xclip
  ```

**问题**：通知不显示
- **解决**：确保桌面环境运行了通知守护进程：
  ```bash
  # GNOME
  ps aux | grep notification-daemon

  # 如果未运行，重启桌面环境或安装：
  sudo apt-get install notification-daemon
  ```

**问题**：TTS/STT质量差
- **解决**：
  - TTS：考虑使用第三方TTS API（如Google Cloud TTS、Amazon Polly等）
  - STT：安装Whisper以获得更好的识别质量

**问题**：CAN总线无法使用
- **解决**：
  - 检查内核是否支持CAN：
    ```bash
    # 检查内核配置
    zgrep CONFIG_CAN /proc/config.gz
    # 或
    grep CONFIG_CAN /boot/config-$(uname -r)
    ```
  - 加载CAN内核模块：
    ```bash
    sudo modprobe can
    sudo modprobe can_raw
    sudo modprobe vcan  # 虚拟CAN接口（用于测试）
    ```
  - 创建虚拟CAN接口（测试用）：
    ```bash
    sudo ip link add dev vcan0 type vcan
    sudo ip link set up vcan0
    ```

**问题**：GPIO无法使用
- **解决**：
  - 检查权限：
    ```bash
    # 将用户添加到gpio组
    sudo usermod -aG gpio $USER
    # 重新登录生效
    ```
  - 检查GPIO设备：
    ```bash
    ls /sys/class/gpio/
    # 或使用libgpiod
    gpiodetect
    gpioinfo
    ```
  - 安装依赖：
    ```bash
    sudo apt-get install gpiod
    ```

---

## 贡献

如果您发现平台兼容性问题或有改进建议，请：
1. 检查本文档是否已包含相关解决方案
2. 提交Issue描述问题
3. 如果可能，提交Pull Request修复问题

---

## 许可证

AgentFramework 采用 AGPL-3.0-or-later 许可证。详见 [LICENSE](../../LICENSE) 文件。
