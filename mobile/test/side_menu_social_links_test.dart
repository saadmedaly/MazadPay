import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';

// Note #3 (client feedback): the drawer's "مشاركة" (share) button and the
// footer social icons (Facebook, TikTok, Instagram, Snapchat) previously had
// no working tap handler at all -- the share button's onPressed was an
// empty no-op, and the social icon Containers had no tap handler/InkWell at
// all. socialPlatformUri is the extracted pure logic (SideMenuDrawer itself
// has no DI seam for a full widget-level launchUrl test, same constraint
// documented elsewhere in this app) that decides what URL, if any, a given
// platform's icon should open -- tested directly here.
//
// Facebook and TikTok use the exact official URLs the client supplied.
// Snapchat and Instagram have no official MazadPay URL anywhere in this
// repository (mobile/backend/web all checked) -- per explicit instruction,
// no URL was invented; socialPlatformUri must return null for both until
// the client supplies real links.
void main() {
  group('socialPlatformUri (Note #3: drawer social icons)', () {
    test('Facebook resolves to the exact client-supplied official page (host + path preserved)', () {
      final uri = socialPlatformUri(SocialPlatform.facebook);
      expect(uri, isNotNull);
      // Uri.tryParse percent-encodes the raw "[0]" in __cft__[0]= into
      // %5B0%5D per RFC 3986 -- this changes the URI's textual form but not
      // its meaning when launched, so the source URL constant (not the
      // parsed Uri's toString()) is what's checked for exact fidelity to
      // the client-supplied link.
      expect(uri!.host, 'web.facebook.com');
      expect(uri.path, '/mazadpay');
      expect(socialPlatformUrls[SocialPlatform.facebook],
          'https://web.facebook.com/mazadpay?__cft__[0]=AZhR63A_FCJiSqEMNQwseVLTbJS_yXsNyEHbRMf1zTttKph3kTH8qGbXVkCtzluqN8vhXAJDlN0ai1HUx2lmNiP9OFCD1maDfZR5eDJR9YancWSU7PxM7Pq-NDIdSFZlkrRyNg2dGYBIQddWdK4Y2_XAnhbgo431Y7qeTjeHXK6iwCPdkXIxu6RfH3d82to&__tn__=%2Cd%3C%2CP-R');
    });

    test('TikTok resolves to the exact client-supplied official profile URL', () {
      final uri = socialPlatformUri(SocialPlatform.tiktok);
      expect(uri, isNotNull);
      expect(uri!.host, 'www.tiktok.com');
      expect(uri.toString(), socialPlatformUrls[SocialPlatform.tiktok]);
    });

    test('Snapchat has no official URL yet -- resolves to null, never a guessed link', () {
      expect(socialPlatformUri(SocialPlatform.snapchat), isNull);
      expect(socialPlatformUrls.containsKey(SocialPlatform.snapchat), isFalse);
    });

    test('Instagram has no official URL yet -- resolves to null, never a guessed link', () {
      expect(socialPlatformUri(SocialPlatform.instagram), isNull);
      expect(socialPlatformUrls.containsKey(SocialPlatform.instagram), isFalse);
    });

    test('Facebook and TikTok URLs are both valid https URIs', () {
      for (final platform in [SocialPlatform.facebook, SocialPlatform.tiktok]) {
        final uri = socialPlatformUri(platform);
        expect(uri!.scheme, 'https');
      }
    });
  });
}
