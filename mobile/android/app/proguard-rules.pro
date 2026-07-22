# Flutter wrapper — obligatoire, sans quoi l'engine Flutter/le plugin registrar plante au démarrage.
-keep class io.flutter.app.** { *; }
-keep class io.flutter.plugin.**  { *; }
-keep class io.flutter.util.**  { *; }
-keep class io.flutter.view.**  { *; }
-keep class io.flutter.**  { *; }
-keep class io.flutter.plugins.**  { *; }
-dontwarn io.flutter.embedding.**

# Firebase Messaging / Analytics — évite la suppression des classes générées par
# réflexion (échecs silencieux de FCM sinon).
-keep class com.google.firebase.** { *; }
-keep class com.google.android.gms.** { *; }
-dontwarn com.google.firebase.**
-dontwarn com.google.android.gms.**

# flutter_secure_storage / shared_preferences — utilisent des accès par réflexion sur
# certains adaptateurs Android.
-keep class androidx.security.crypto.** { *; }

# json_annotation / modèles sérialisés (Auction, Wallet, Transaction, etc.) — évite que
# R8 renomme les champs utilisés par la (dé)sérialisation JSON manuelle.
-keepattributes *Annotation*
-keepattributes Signature
-keepattributes EnclosingMethod
-keepattributes InnerClasses

# Play Core (mode de démarrage différé de Flutter) — supprime les avertissements
# inoffensifs liés aux classes optionnelles absentes du classpath.
-dontwarn com.google.android.play.core.**
