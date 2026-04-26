import 'package:geolocator/geolocator.dart';
import 'package:geocoding/geocoding.dart';

/// Service pour la détection de localisation GPS
class LocationService {
  /// Vérifier et demander les permissions de localisation
  static Future<bool> handlePermission() async {
    bool serviceEnabled;
    LocationPermission permission;

    // Vérifier si le service GPS est activé
    serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      return false;
    }

    // Vérifier la permission
    permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        return false;
      }
    }

    if (permission == LocationPermission.deniedForever) {
      return false;
    }

    return true;
  }

  /// Obtenir la position actuelle
  static Future<Position?> getCurrentPosition() async {
    try {
      final hasPermission = await handlePermission();
      if (!hasPermission) return null;

      return await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
      );
    } catch (e) {
      return null;
    }
  }

  /// Convertir les coordonnées en adresse (ville + pays)
  static Future<LocationData?> getAddressFromCoordinates(
    double latitude,
    double longitude,
  ) async {
    try {
      final placemarks = await placemarkFromCoordinates(latitude, longitude);
      
      if (placemarks.isNotEmpty) {
        final place = placemarks.first;
        return LocationData(
          city: place.locality ?? place.subAdministrativeArea ?? place.administrativeArea,
          country: place.country,
          countryCode: place.isoCountryCode,
          street: place.street,
          latitude: latitude,
          longitude: longitude,
        );
      }
      return null;
    } catch (e) {
      return null;
    }
  }

  /// Détecter la localisation complète en une seule méthode
  static Future<LocationData?> detectLocation() async {
    final position = await getCurrentPosition();
    if (position == null) return null;

    return await getAddressFromCoordinates(
      position.latitude,
      position.longitude,
    );
  }

  /// Fallback pour Nouakchott (capitale par défaut)
  static LocationData getDefaultLocation() {
    return LocationData(
      city: 'نواكشوط',
      country: 'Mauritania',
      countryCode: 'MR',
      latitude: 18.0735,
      longitude: -15.9582,
    );
  }
}

/// Modèle de données de localisation
class LocationData {
  final String? city;
  final String? country;
  final String? countryCode;
  final String? street;
  final double latitude;
  final double longitude;

  LocationData({
    this.city,
    this.country,
    this.countryCode,
    this.street,
    required this.latitude,
    required this.longitude,
  });

  String get displayName {
    if (city != null && city!.isNotEmpty) {
      return city!;
    }
    if (country != null && country!.isNotEmpty) {
      return country!;
    }
    return 'Localisation inconnue';
  }

  Map<String, dynamic> toJson() {
    return {
      'city': city,
      'country': country,
      'countryCode': countryCode,
      'street': street,
      'latitude': latitude,
      'longitude': longitude,
    };
  }

  factory LocationData.fromJson(Map<String, dynamic> json) {
    return LocationData(
      city: json['city'],
      country: json['country'],
      countryCode: json['countryCode'],
      street: json['street'],
      latitude: json['latitude'] ?? 0.0,
      longitude: json['longitude'] ?? 0.0,
    );
  }
}
