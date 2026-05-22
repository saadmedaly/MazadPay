import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/pages/delivery_details_page.dart';

class ServicesPage extends StatelessWidget {
  const ServicesPage({super.key});

  // Banner image URL from admin panel (null = use local asset)
  static const String? _bannerImageUrl = null;

  static const List<_ServiceItem> _services = [
    _ServiceItem('كورس',           Icons.local_taxi,          Color(0xFFFFF8E1), Colors.orange),
    _ServiceItem('توصيل',          Icons.delivery_dining,     Color(0xFFE8F5E9), Color(0xFF2E7D32)),
    _ServiceItem('نقل البضائع',    Icons.local_shipping,      Color(0xFFFFF3E0), Colors.deepOrange),
    _ServiceItem('كورس عبر المدن', Icons.directions_bus,      Color(0xFFE8F5E9), Colors.green),
    _ServiceItem('توصيل طعام',     Icons.restaurant,          Color(0xFFFCE4EC), Colors.red),
    _ServiceItem('توصيل أدوية',    Icons.medical_services,    Color(0xFFE3F2FD), Colors.blue),
    _ServiceItem('شحن من خارج',   Icons.flight,              Color(0xFFE8EAF6), Color(0xFF3949AB)),
    _ServiceItem('رافعة سيارة',    Icons.car_repair,          Color(0xFFFFF8E1), Colors.amber),
    _ServiceItem('شاحنة ماء',      Icons.local_shipping,      Color(0xFFE3F2FD), Color(0xFF0277BD)),
    _ServiceItem('توصيل مجاري',    Icons.water_damage,        Color(0xFFE8F5E9), Color(0xFF2E7D32)),
    _ServiceItem('نقل أثاث',       Icons.chair,               Color(0xFFF3E5F5), Color(0xFF8E24AA)),
    _ServiceItem('توصيل أسماك ولحوم', Icons.set_meal,         Color(0xFFE0F7FA), Color(0xFF00838F)),
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
                  onTap: svc.title == 'توصيل'
                      ? () => Navigator.push(
                            context,
                            MaterialPageRoute(builder: (_) => const DeliveryDetailsPage()),
                          )
                      : null,
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
    return InkWell(
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
