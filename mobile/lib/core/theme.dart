import 'package:flutter/material.dart';


class AppTheme {
  static const Color primaryColor = Color(0xFF0081FF);
  static const Color accentColor = Color(0xFF0081FF);
  static const Color backgroundColor = Color(0xFFFBFBFB);
  static const Color darkBackgroundColor = Color(0xFF121212);

  static ThemeData lightTheme = ThemeData(
    useMaterial3: true,
    primaryColor: primaryColor,
    scaffoldBackgroundColor: backgroundColor,
    colorScheme: ColorScheme.fromSeed(
      seedColor: primaryColor,
      primary: primaryColor,
      surface: backgroundColor,
    ),
    textTheme: ThemeData.light().textTheme.apply(
      fontFamily: 'Poppins',
      fontFamilyFallback: const ['Cairo'],
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: Colors.transparent,
      elevation: 0,
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: primaryColor,
        foregroundColor: Colors.white,
        minimumSize: const Size(double.infinity, 48),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        elevation: 0,
        textStyle: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, fontFamilyFallback: ['Cairo']),
      ),
    ),
  );

  static ThemeData darkTheme = ThemeData(
    useMaterial3: true,
    primaryColor: primaryColor,
    scaffoldBackgroundColor: darkBackgroundColor,
    colorScheme: ColorScheme.fromSeed(
      brightness: Brightness.dark,
      seedColor: primaryColor,
      primary: primaryColor,
      surface: const Color(0xFF1D1D1D),
    ),
    textTheme: ThemeData.dark().textTheme.apply(
      fontFamily: 'Poppins',
      fontFamilyFallback: const ['Cairo'],
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: Colors.transparent,
      elevation: 0,
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: primaryColor,
        foregroundColor: Colors.white,
        minimumSize: const Size(double.infinity, 48),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        elevation: 0,
        textStyle: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
      ),
    ),
  );

}
