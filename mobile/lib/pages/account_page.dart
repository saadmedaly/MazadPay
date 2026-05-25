import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'deposit_page.dart';
import 'account_profile_page.dart';
import 'my_auctions_page.dart';
import 'favorites_page.dart';
import 'my_winnings_page.dart';
import 'withdraw_page.dart';
import '../services/wallet_api.dart';

class AccountPage extends StatefulWidget {
  const AccountPage({super.key});

  @override
  State<AccountPage> createState() => _AccountPageState();
}

class _AccountPageState extends State<AccountPage> {
  bool _showBalance = false;
  final WalletApi _walletApi = WalletApi();
  double _balance = 0.0;

  @override
  void initState() {
    super.initState();
    _loadBalance();
  }

  Future<void> _loadBalance() async {
    try {
      final response = await _walletApi.getBalance();

      if (!mounted) return;
      setState(() {
        if (response.success && response.data != null) {
          _balance = (response.data!['balance'] ?? 0).toDouble();
        }
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(AppLocalizations.of(context)!.error_loading_balance),
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 16),

              // Insurance Card
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFF0084FF), Color(0xFF0055FF)],
                    begin: AlignmentDirectional.centerEnd,
                    end: AlignmentDirectional.centerStart,
                  ),
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: primaryBlue.withValues(alpha: 0.3),
                      blurRadius: 10,
                      offset: const Offset(0, 5),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          AppLocalizations.of(context)!.text_20,
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Icon(Icons.more_horiz, color: Colors.white),
                      ],
                    ),
                    const SizedBox(height: 24),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              AppLocalizations.of(context)!.text_21,
                              style: TextStyle(
                                color: Colors.white70,
                                fontSize: 12,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              _showBalance
                                  ? '••••••'
                                  : '${_balance.toStringAsFixed(2)} MRU',
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                              ),
                              textDirection: TextDirection.ltr,
                            ),
                          ],
                        ),
                        IconButton(
                          icon: Icon(
                            _showBalance
                                ? Icons.visibility_off_outlined
                                : Icons.visibility_outlined,
                            color: Colors.white,
                          ),
                          onPressed: () {
                            setState(() {
                              _showBalance = !_showBalance;
                            });
                          },
                        ),
                      ],
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),
              Text(
                AppLocalizations.of(context)!.text_22,
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
              ),
              const SizedBox(height: 16),

              // Deposit Now Button (Red)
              InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const DepositPage(),
                    ),
                  );
                },
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 20,
                  ),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFFE50000), Color(0xFFFF4040)],
                      begin: AlignmentDirectional.centerEnd,
                      end: AlignmentDirectional.centerStart,
                    ),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Icon(
                        Icons.arrow_back_ios,
                        color: Colors.white,
                        size: 20,
                      ),
                      Flexible(
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Flexible(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.end,
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Text(
                                    AppLocalizations.of(context)!.text_23,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 18,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    AppLocalizations.of(context)!.text_24,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 12,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(width: 16),
                            const Icon(
                              Icons.account_balance_wallet,
                              color: Colors.white,
                              size: 40,
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 16),

              // Retrieve Insurance Button (Green)
              InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const WithdrawPage(),
                    ),
                  );
                },
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 20,
                  ),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFF00AA00), Color(0xFF33CC33)],
                      begin: AlignmentDirectional.centerEnd,
                      end: AlignmentDirectional.centerStart,
                    ),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Icon(
                        Icons.arrow_back_ios,
                        color: Colors.white,
                        size: 20,
                      ),
                      Flexible(
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Flexible(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.end,
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Text(
                                    AppLocalizations.of(context)!.text_25,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 18,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    AppLocalizations.of(context)!.text_25,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 12,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(width: 16),
                            const Icon(
                              Icons.account_balance_wallet,
                              color: Colors.white,
                              size: 40,
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 32),
              Text(
                AppLocalizations.of(context)!.text_26,
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
              ),
              const SizedBox(height: 16),

              // Activity Grid
              Row(
                children: [
                  Expanded(
                    child: _buildActivityItem(
                      icon: Icons.gavel,
                      title: AppLocalizations.of(context)!.text_27,
                      isDarkMode: isDarkMode,
                      onTap: () => Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => const MyAuctionsPage(),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _buildActivityItem(
                      icon: Icons.favorite_border,
                      title: AppLocalizations.of(context)!.text_28,
                      iconColor: Colors.red,
                      isDarkMode: isDarkMode,
                      onTap: () => Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => const FavoritesPage(),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _buildActivityItem(
                      icon: Icons.emoji_events_outlined,
                      title: AppLocalizations.of(context)!.text_29,
                      isDarkMode: isDarkMode,
                      onTap: () => Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => const MyWinningsPage(),
                        ),
                      ),
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 24),
              Center(
                child: Column(
                  children: [
                    Text(
                      AppLocalizations.of(context)!.text_30,
                      style: TextStyle(color: Colors.grey[600], fontSize: 12),
                    ),
                    const SizedBox(height: 4),
                    InkWell(
                      onTap: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) => const AccountProfilePage(),
                          ),
                        );
                      },
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.person_outline, size: 16),
                          SizedBox(width: 4),
                          Text(
                            AppLocalizations.of(context)!.text_31,
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 12,
                              color: Color(0xFF0084FF),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 80), // Space for bottom nav
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildActivityItem({
    required IconData icon,
    required String title,
    Color? iconColor,
    required bool isDarkMode,
    VoidCallback? onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 8),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.withValues(alpha: 0.2)),
          boxShadow: [
            if (!isDarkMode)
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.02),
                blurRadius: 5,
                offset: const Offset(0, 2),
              ),
          ],
        ),
        child: Column(
          children: [
            Icon(
              icon,
              color: iconColor ?? (isDarkMode ? Colors.white : Colors.black87),
              size: 28,
            ),
            const SizedBox(height: 12),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
            ),
          ],
        ),
      ),
    );
  }
}
