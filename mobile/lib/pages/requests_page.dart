import 'package:flutter/material.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/services/request_api.dart';

class RequestsPage extends StatefulWidget {
  const RequestsPage({super.key});

  @override
  State<RequestsPage> createState() => _RequestsPageState();
}

class _RequestsPageState extends State<RequestsPage> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final RequestApi _requestApi = RequestApi();
  bool _isLoading = true;
  List<dynamic> _auctionRequests = [];
  List<dynamic> _bannerRequests = [];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _loadRequests();
  }

  Future<void> _loadRequests() async {
    setState(() => _isLoading = true);
    try {
      final auctionResp = await _requestApi.getMyAuctionRequests();
      final bannerResp = await _requestApi.getMyBannerRequests();

      if (mounted) {
        setState(() {
          if (auctionResp.success && auctionResp.data != null) {
            _auctionRequests = auctionResp.data!['requests'] ?? [];
          }
          if (bannerResp.success && bannerResp.data != null) {
            _bannerRequests = bannerResp.data!['requests'] ?? [];
          }
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        title: Text(localizations.text_401),
        bottom: TabBar(
          controller: _tabController,
          tabs: [
            Tab(text: localizations.text_90), // "Create ad" (as proxy for auction request)
            const Tab(text: 'Bannières'),
          ],
        ),
      ),
      body: _isLoading 
        ? const Center(child: CircularProgressIndicator())
        : TabBarView(
            controller: _tabController,
            children: [
              _buildRequestList(_auctionRequests, isDarkMode),
              _buildRequestList(_bannerRequests, isDarkMode),
            ],
          ),
    );
  }

  Widget _buildRequestList(List<dynamic> requests, bool isDarkMode) {
    if (requests.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.inventory_2_outlined, size: 64, color: Colors.grey[400]),
            const SizedBox(height: 16),
            const Text('Aucune demande trouvée', style: TextStyle(color: Colors.grey)),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadRequests,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: requests.length,
        itemBuilder: (context, index) {
          final req = requests[index];
          return _buildRequestCard(req, isDarkMode);
        },
      ),
    );
  }

  Widget _buildRequestCard(dynamic req, bool isDarkMode) {
    final status = req['status']?.toString().toLowerCase() ?? 'pending';
    Color statusColor = Colors.orange;
    if (status == 'approved') statusColor = Colors.green;
    if (status == 'rejected') statusColor = Colors.red;

    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: Text(
                    req['title']?.toString() ?? 'Sans titre',
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    status.toUpperCase(),
                    style: TextStyle(color: statusColor, fontSize: 10, fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              req['description']?.toString() ?? '',
              style: TextStyle(color: Colors.grey[600], fontSize: 14),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  _formatDate(req['created_at']?.toString() ?? ''),
                  style: TextStyle(color: Colors.grey[400], fontSize: 12),
                ),
                if (req['amount'] != null)
                  Text(
                    '${req['amount']} MRU',
                    style: const TextStyle(fontWeight: FontWeight.bold, color: Color(0xFF0081FF)),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  String _formatDate(String dateStr) {
    if (dateStr.isEmpty) return '';
    try {
      final date = DateTime.parse(dateStr);
      return '${date.day}/${date.month}/${date.year}';
    } catch (e) {
      return dateStr;
    }
  }
}
