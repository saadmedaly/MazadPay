import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/utils/whatsapp_launcher.dart';

class ServicesPage extends StatelessWidget {
  const ServicesPage({super.key});

  // Banner image URL from admin panel (null = use local asset)
  static const String? _bannerImageUrl = null;

  // Note #4 (client feedback): reduced from the prior 12 services to
  // exactly these 5, per the client's authoritative reference image.
  // Same visual style/layout (icon + colored background + label) --
  // nothing else about the grid/card design changed, only the list
  // contents.
  static const List<_ServiceItem> _services = [
    _ServiceItem('نقل البضائع', Icons.local_shipping, Color(0xFFFFF3E0), Colors.deepOrange),
    _ServiceItem('توصيل',       Icons.delivery_dining, Color(0xFFE8F5E9), Color(0xFF2E7D32)),
    _ServiceItem('رافعة سيارة', Icons.car_repair,     Color(0xFFFFF8E1), Colors.amber),
    _ServiceItem('نقل أثاث',    Icons.chair,          Color(0xFFF3E5F5), Color(0xFF8E24AA)),
    _ServiceItem('شحن من خارج', Icons.flight,         Color(0xFFE8EAF6), Color(0xFF3949AB)),
  ];

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Promo Banner ──────────────────────────────────────────────
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: _bannerImageUrl != null && _bannerImageUrl!.isNotEmpty
                  ? Image.network(
                      _bannerImageUrl!,
                      width: double.infinity,
                      fit: BoxFit.contain,
                      errorBuilder: (_, __, ___) => _localBanner(),
                    )
                  : _localBanner(),
            ),
          ),

          // ── Section Title ─────────────────────────────────────────────
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
            child: Text(
              AppLocalizations.of(context)!.text_297,
              style: const TextStyle(
                fontFamily: 'Plus Jakarta Sans',
                fontSize: 22,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),

          // ── Services Grid ─────────────────────────────────────────────
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 3,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 0.9,
              ),
              itemCount: _services.length,
              itemBuilder: (context, index) {
                final svc = _services[index];
                return _buildServiceCard(
                  context,
                  svc,
                  // Note #4: tapping any of the 5 delivery-service cards
                  // opens MazadPay's own WhatsApp (the same reliable
                  // wa.me + whatsapp:// fallback strategy already fixed in
                  // My Winnings), prefilled with a message naming the
                  // tapped service -- not the per-service navigation this
                  // grid used to have before client feedback A5 removed it.
                  onTap: () => launchMazadPayWhatsApp(
                    context,
                    'مرحباً، أرغب في الاستفسار عن خدمة "${svc.title}".',
                  ),
                );
              },
            ),
          ),

          const SizedBox(height: 100),
        ],
      ),
    );
  }

  Widget _localBanner() {
    return Image.asset(
      'assets/khedemat.jpg',
      width: double.infinity,
      fit: BoxFit.contain,
      errorBuilder: (_, __, ___) => Container(
        height: 160,
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [Color(0xFF0084FF), Color(0xFF0055FF)],
            begin: AlignmentDirectional.centerEnd,
            end: AlignmentDirectional.centerStart,
          ),
          borderRadius: BorderRadius.circular(16),
        ),
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.all(20),
        child: const Text(
          'اربح وقتك',
          style: TextStyle(color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
        ),
      ),
    );
  }

  Widget _buildServiceCard(BuildContext context, _ServiceItem svc, {VoidCallback? onTap}) {
    return Material(
      type: MaterialType.transparency,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: Container(
          decoration: BoxDecoration(
            color: svc.bgColor,
            borderRadius: BorderRadius.circular(16),
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(svc.icon, size: 40, color: svc.iconColor),
              const SizedBox(height: 10),
              Text(
                svc.title,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontFamily: 'Plus Jakarta Sans',
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ServiceItem {
  final String title;
  final IconData icon;
  final Color bgColor;
  final Color iconColor;
  const _ServiceItem(this.title, this.icon, this.bgColor, this.iconColor);
}
