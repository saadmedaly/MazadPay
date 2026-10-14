import 'package:flutter/material.dart';

/// Days/hours/minutes/seconds breakdown of a countdown [Duration], as used
/// by the day/hour/minute/second boxes on the auction details page. Pure and
/// side-effect-free so it's directly unit-testable. Negative durations are
/// clamped to all-zero (the countdown Timer itself clamps to Duration.zero
/// before this is ever called, but this stays safe on its own too).
class CountdownParts {
  final int days;
  final int hours;
  final int minutes;
  final int seconds;

  const CountdownParts({
    required this.days,
    required this.hours,
    required this.minutes,
    required this.seconds,
  });

  factory CountdownParts.fromDuration(Duration duration) {
    if (duration.isNegative) {
      return const CountdownParts(days: 0, hours: 0, minutes: 0, seconds: 0);
    }
    return CountdownParts(
      days: duration.inDays,
      hours: duration.inHours % 24,
      minutes: duration.inMinutes % 60,
      seconds: duration.inSeconds % 60,
    );
  }
}

class TimeUtils {
  static String formatDuration(BuildContext context, Duration duration) {
    final locale = Localizations.localeOf(context).languageCode;
    
    if (duration.isNegative) {
      return locale == 'ar' ? 'انتهى' : (locale == 'fr' ? 'Terminé' : 'Ended');
    }

    if (duration.inDays > 0) {
      final days = duration.inDays;
      final hours = duration.inHours % 24;
      final d = locale == 'ar' ? 'ي' : 'd';
      final h = locale == 'ar' ? 'س' : 'h';
      final r = locale == 'ar' ? 'متبقي' : (locale == 'fr' ? 'restants' : 'remaining');
      return '$days$d $hours$h $r';
    } else if (duration.inHours > 0) {
      final hours = duration.inHours;
      final minutes = duration.inMinutes % 60;
      final h = locale == 'ar' ? 'س' : 'h';
      final m = locale == 'ar' ? 'د' : 'm';
      final r = locale == 'ar' ? 'متبقي' : (locale == 'fr' ? 'restants' : 'remaining');
      return '$hours$h $minutes$m $r';
    } else {
      final minutes = duration.inMinutes;
      final m = locale == 'ar' ? 'د' : 'm';
      final r = locale == 'ar' ? 'متبقي' : (locale == 'fr' ? 'restants' : 'remaining');
      return '$minutes$m $r';
    }
  }
}
