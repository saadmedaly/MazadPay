import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/home_page.dart';

// Client feedback: Bug F. On cold start (first launch, or reopened after a
// full close), the Home page briefly showed an old commercial ad image
// (assets/announcement.png) before the real server banner appeared. Root
// cause: the Hero Banner carousel's Builder rendered the SAME
// _buildBannerCard (a hardcoded Image.asset('assets/announcement.png'))
// for both the "banner API still loading" state AND the "no banners"
// state -- and _isLoadingBanners starts true, so that old ad was exactly
// what rendered on the very first frame, every time, until the API
// resolved. It was never a real connectivity issue.
//
// bannerSlideMode is the extracted, pure decision logic that replaces the
// old inline ternary chain -- tested directly here. The widget-level fix
// (home_page.dart) now renders a neutral grey placeholder (with a spinner
// only while loading) for both the `loading` and `empty` modes, and never
// references assets/announcement.png in that path at all anymore.
void main() {
  group('bannerSlideMode (Bug F cold-start flash fix)', () {
    test('while the banner API is still loading, mode is loading (never data)', () {
      final mode = bannerSlideMode(isLoadingBanners: true, banners: []);
      expect(mode, BannerSlideMode.loading);
    });

    test('loading is true even if banners already has stale content from a previous state', () {
      // Guards against ever reintroducing a path where a hardcoded/stale
      // list is mistaken for real data while a fetch is still in flight.
      final mode = bannerSlideMode(
        isLoadingBanners: true,
        banners: [
          {'id': 'stale', 'image_url': 'https://example.com/stale.png'},
        ],
      );
      expect(mode, BannerSlideMode.loading);
    });

    test('after loading finishes with zero active banners, mode is empty (neutral placeholder, not the old ad)', () {
      final mode = bannerSlideMode(isLoadingBanners: false, banners: []);
      expect(mode, BannerSlideMode.empty);
    });

    test('after loading finishes with one banner, mode is data', () {
      final mode = bannerSlideMode(
        isLoadingBanners: false,
        banners: [
          {'id': 'b1', 'image_url': 'https://example.com/banner1.png'},
        ],
      );
      expect(mode, BannerSlideMode.data);
    });

    test('after loading finishes with multiple banners, mode is data (PageView rotation applies)', () {
      final mode = bannerSlideMode(
        isLoadingBanners: false,
        banners: [
          {'id': 'b1', 'image_url': 'https://example.com/banner1.png'},
          {'id': 'b2', 'image_url': 'https://example.com/banner2.png'},
          {'id': 'b3', 'image_url': 'https://example.com/banner3.png'},
        ],
      );
      expect(mode, BannerSlideMode.data);
    });

    test('loading takes precedence over empty/data whenever isLoadingBanners is true', () {
      // Mirrors the real _loadBanners flow: a cache hit sets _isLoadingBanners
      // = false immediately, but until either the cache or the API responds,
      // it stays true regardless of what _banners currently holds.
      expect(
        bannerSlideMode(isLoadingBanners: true, banners: []),
        isNot(BannerSlideMode.data),
      );
    });
  });
}
