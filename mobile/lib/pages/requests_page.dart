import 'package:flutter/material.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/create_ad_form_page.dart';
import 'package:mezadpay/pages/create_banner_request_page.dart';
import 'package:mezadpay/services/request_api.dart';
import 'package:mezadpay/utils/money_formatter.dart';

class RequestsPage extends StatefulWidget {
  const RequestsPage({super.key});

  @override
  State<RequestsPage> createState() => _RequestsPageState();
}

/// Pure status-badge classification for a request card, extracted for unit
/// testability. [status] is the raw (lower-cased upstream) request status;
/// [isAuctionTab] mirrors `_tabController.index == 0` — resubmission is only
/// offered for auction requests (banners have no PUT /requests/... update
/// endpoint), and only in the 'draft' or 'rejected' states.
class RequestStatusInfo {
  final String statusKey; // one of: approved, rejected, draft, pending
  final bool canEditResubmit;

  const RequestStatusInfo({required this.statusKey, required this.canEditResubmit});
}

RequestStatusInfo classifyRequestStatus(String? rawStatus, {required bool isAuctionTab}) {
  final status = rawStatus?.toLowerCase() ?? 'pending';
  String statusKey;
  switch (status) {
    case 'approved':
    case 'active':
      statusKey = 'approved';
      break;
    case 'rejected':
      statusKey = 'rejected';
      break;
    case 'draft':
      statusKey = 'draft';
      break;
    default:
      statusKey = 'pending';
  }
  final canEditResubmit = isAuctionTab && (status == 'draft' || status == 'rejected');
  return RequestStatusInfo(statusKey: statusKey, canEditResubmit: canEditResubmit);
}

class _RequestsPageState extends State<RequestsPage>
    with SingleTickerProviderStateMixin {
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
    final lang = Localizations.localeOf(context).languageCode;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    String bannerTabTitle = 'Bannières';
    String createBannerLabel = 'Demander une bannière';
    if (lang == 'ar') {
      bannerTabTitle = 'الإعلانات';
      createBannerLabel = 'طلب إعلان';
    } else if (lang == 'en') {
      bannerTabTitle = 'Banners';
      createBannerLabel = 'Request a Banner';
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(localizations.text_401),
        bottom: TabBar(
          controller: _tabController,
          tabs: [
            Tab(
              text: localizations.text_90,
            ), // "Create ad" (as proxy for auction request)
            Tab(text: bannerTabTitle),
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
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          if (_tabController.index == 0) {
            Navigator.of(context)
                .push(
                  MaterialPageRoute(
                    builder: (context) => const CreateAdFormPage(),
                  ),
                )
                .then((_) => _loadRequests());
          } else {
            Navigator.of(context)
                .push(
                  MaterialPageRoute(
                    builder: (context) => const CreateBannerRequestPage(),
                  ),
                )
                .then((_) => _loadRequests());
          }
        },
        icon: const Icon(Icons.add, color: Colors.white),
        label: Text(
          _tabController.index == 0 ? localizations.text_90 : createBannerLabel,
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
          ),
        ),
        backgroundColor: const Color(0xFF0081FF),
      ),
    );
  }

  Widget _buildRequestList(List<dynamic> requests, bool isDarkMode) {
    final lang = Localizations.localeOf(context).languageCode;
    String emptyMsg = 'Aucune demande trouvée';
    if (lang == 'ar') emptyMsg = 'لم يتم العثور على طلبات';
    if (lang == 'en') emptyMsg = 'No requests found';

    if (requests.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.inventory_2_outlined, size: 64, color: Colors.grey[400]),
            const SizedBox(height: 16),
            Text(emptyMsg, style: const TextStyle(color: Colors.grey)),
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

  String _getLocalized(dynamic req, String field) {
    final lang = Localizations.localeOf(context).languageCode;
    String noTitle = 'Sans titre';
    if (lang == 'ar') noTitle = 'بدون عنوان';
    if (lang == 'en') noTitle = 'No title';

    // Try field_lang, then fallback to field_ar, then field_en, then field
    final val =
        req['${field}_$lang'] ??
        req['${field}_ar'] ??
        req['${field}_en'] ??
        req[field];

    if (val == null || val.toString().isEmpty) {
      return field == 'title' ? noTitle : '';
    }
    return val.toString();
  }

  Widget _buildRequestCard(dynamic req, bool isDarkMode) {
    final l10n = AppLocalizations.of(context)!;
    final rawStatus = req['status']?.toString();
    final status = rawStatus?.toLowerCase() ?? 'pending';
    // Le bouton "Modifier et renvoyer" n'est pertinent que pour les demandes
    // d'enchères (auction_requests) en brouillon ou rejetées — PUT
    // /requests/auctions/:id n'existe pas pour les bannières, et le backend
    // n'autorise cette mise à jour que sur ces deux statuts (voir
    // RequestService.UpdateAuctionRequest).
    final info = classifyRequestStatus(rawStatus, isAuctionTab: _tabController.index == 0);
    final canEditResubmit = info.canEditResubmit;

    Color statusColor;
    String statusText;
    switch (info.statusKey) {
      case 'approved':
        statusColor = Colors.green;
        statusText = l10n.status_approved;
        break;
      case 'rejected':
        statusColor = Colors.red;
        statusText = l10n.status_rejected;
        break;
      case 'draft':
        statusColor = Colors.grey;
        statusText = l10n.status_draft;
        break;
      default:
        statusColor = Colors.orange;
        statusText = l10n.status_pending;
    }

    final adminNotes = req['admin_notes']?.toString();

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
                    _getLocalized(req, 'title'),
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: statusColor.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    statusText,
                    style: TextStyle(
                      color: statusColor,
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              _getLocalized(req, 'description'),
              style: TextStyle(color: Colors.grey[600], fontSize: 14),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            if (status == 'rejected' &&
                adminNotes != null &&
                adminNotes.isNotEmpty) ...[
              const SizedBox(height: 8),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: Colors.red.withOpacity(0.08),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Icon(Icons.info_outline, color: Colors.red, size: 16),
                    const SizedBox(width: 6),
                    Expanded(
                      child: Text(
                        '${l10n.label_admin_notes}: $adminNotes',
                        style: const TextStyle(color: Colors.red, fontSize: 12),
                      ),
                    ),
                  ],
                ),
              ),
            ],
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
                    MoneyFormatter.format(
                      num.tryParse(req['amount'].toString()) ?? 0,
                      req['currency_code']?.toString(),
                    ),
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      color: Color(0xFF0081FF),
                    ),
                  ),
              ],
            ),
            if (canEditResubmit) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: () {
                    Navigator.of(context)
                        .push(
                          MaterialPageRoute(
                            builder: (context) => CreateAdFormPage(
                              existingRequest: req as Map<String, dynamic>,
                            ),
                          ),
                        )
                        .then((_) => _loadRequests());
                  },
                  icon: const Icon(Icons.edit_outlined, size: 18),
                  label: Text(l10n.action_edit_resubmit),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: const Color(0xFF0081FF),
                    side: const BorderSide(color: Color(0xFF0081FF)),
                  ),
                ),
              ),
            ],
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
