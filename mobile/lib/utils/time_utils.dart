import 'package:flutter/material.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

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
