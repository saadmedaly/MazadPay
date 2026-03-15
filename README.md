# MazadPay 🚀

MazadPay is a premium auction and digital payment application built with Flutter. It provides a seamless, high-performance experience for users to participate in auctions, manage their digital wallet, and handle payments in multiple languages.

## ✨ Key Features

### 🌍 Advanced Internationalization
- **Multi-language Support**: Fully localized in **Arabic**, **French**, and **English**.
- **Dynamic RTL/LTR**: Automatic layout switching based on the selected locale.
- **In-app Language Switcher**: Persistent locale management using the `Provider` pattern.

### 🍱 User Experience & Design
- **Consistent Navigation**: A universal top menu (CustomAppBar) providing easy access to the side menu drawer from any primary screen.
- **Premium UI**: Modern dark mode support, smooth micro-animations, and glassmorphism elements.
- **Responsive Layout**: Optimized for various screen sizes and orientations.

### 💰 Auction & Payment Features
- **Bidding System**: Real-time auction browsing and bidding interfaces.
- **Digital Wallet**: Manage deposits, view balance (with privacy toggle), and track financial activity.
- **Secure Payments**: Integrated flows for bank transfers (Bankily, Masrvi, Sedad, Click) with receipt upload verification.

## 🛠️ Tech Stack
- **Framework**: [Flutter](https://flutter.dev)
- **State Management**: [Provider](https://pub.dev/packages/provider)
- **Localization**: `flutter_localizations` & `intl` (ARB files)
- **Typography**: Spline Sans & Google Fonts (Inter, Cairo)
- **Icons**: Lucide Icons & Material Icons

## 🚀 Getting Started

### Prerequisites
- Flutter SDK (Latest Stable)
- Android Studio / VS Code
- Git

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/saadmedaly/MazadPay.git
   ```
2. Install dependencies:
   ```bash
   flutter pub get
   ```
3. Generate localization files:
   ```bash
   flutter gen-l10n
   ```
4. Run the app:
   ```bash
   flutter run
   ```

## 🏗️ Project Structure
- `lib/l10n/`: Translation files (.arb) and generated localization classes.
- `lib/pages/`: Main application screens (Home, Bidding, Account, etc.).
- `lib/widgets/`: Reusable UI components (CustomAppBar, SideMenuDrawer, etc.).
- `lib/core/`: Application themes, constants, and utilities.
- `lib/providers/`: State management logic for locale and user data.

## 🏆 Code Quality
- **0 Analysis Issues**: Maintainer-ready code with a perfectly clean `flutter analyze` report.
- **Modern Syntax**: Uses the latest Flutter features (e.g., `.withValues()` for colors).
- **Consistent Style**: Adheres strictly to Dart's official naming and formatting conventions.

---
Developed with ❤️ by the MazadPay Saad Meiloud.
