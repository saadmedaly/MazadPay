import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'dart:convert';

import 'auction_details_page.dart';
import 'auction_history_page.dart';
import 'create_ad_form_page.dart';
import '../services/auction_api.dart';
import '../services/cache_service.dart';
import '../widgets/app_modals.dart';

class MyAuctionsPage extends ConsumerStatefulWidget {
  const MyAuctionsPage({super.key});

  @override
  ConsumerState<MyAuctionsPage> createState() => _MyAuctionsPageState();
}

class _MyAuctionsPageState extends ConsumerState<MyAuctionsPage> {
  final AuctionApi _auctionApi = AuctionApi();
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  
  List<Map<String, dynamic>> _myAuctions = [];
  List<Map<String, dynamic>> _filteredAuctions = [];
  bool _isLoading = true;
  bool _isLoadingMore = false;
  String? _error;
  
  // Pagination
  int _currentPage = 1;
  final int _limit = 20;
  int _totalAuctions = 0;
  bool _hasMore = true;
  
  // Filtres
  String _searchQuery = '';
  String? _selectedStatus;

  @override
  void initState() {
    super.initState();
    _loadMyAuctions();
    _scrollController.addListener(_onScroll);
  }
  
  @override
  void dispose() {
    _searchController.dispose();
    _scrollController.dispose();
    super.dispose();
  }
  
  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent * 0.8) {
      if (!_isLoadingMore && _hasMore && !_isLoading) {
        _loadMoreAuctions();
      }
    }
  }
  
  void _filterAuctions() {
    if (_searchQuery.isEmpty && _selectedStatus == null) {
      setState(() {
        _filteredAuctions = List.from(_myAuctions);
      });
      return;
    }
    
    setState(() {
      _filteredAuctions = _myAuctions.where((auction) {
        // Filtre par recherche texte
        final titleAr = auction['title_ar']?.toString().toLowerCase() ?? '';
        final titleFr = auction['title_fr']?.toString().toLowerCase() ?? '';
        final titleEn = auction['title_en']?.toString().toLowerCase() ?? '';
        final category = auction['category']?.toString().toLowerCase() ?? '';
        final lotNumber = auction['lot_number']?.toString().toLowerCase() ?? '';
        final query = _searchQuery.toLowerCase();
        
        bool matchesSearch = _searchQuery.isEmpty ||
            titleAr.contains(query) ||
            titleFr.contains(query) ||
            titleEn.contains(query) ||
            category.contains(query) ||
            lotNumber.contains(query);
        
         bool matchesStatus = _selectedStatus == null ||
            auction['status']?.toString().toLowerCase() == _selectedStatus!.toLowerCase();
        
        return matchesSearch && matchesStatus;
      }).toList();
    });
  }

  Future<void> _loadMyAuctions() async {
    try {
      setState(() {
        _currentPage = 1;
        _hasMore = true;
      });
      
       final cachedMyAuctions = await CacheService.instance.getCachedMyAuctions();
      final isCacheValid = await CacheService.instance.isMyAuctionsCacheValid();

      if (cachedMyAuctions != null && isCacheValid) {
        List<dynamic> auctionList = [];
        auctionList = (cachedMyAuctions['auctions'] ?? cachedMyAuctions['data'] ?? []) as List<dynamic>;
        
        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _myAuctions = auctionList.map((item) => item as Map<String, dynamic>).toList();
          _filteredAuctions = List.from(_myAuctions);
        });
      }

      // === CHARGER DEPUIS L'API EN ARRIÈRE-PLAN ===
      debugPrint('Chargement des encheres depuis API...');
      final response = await _auctionApi.getMyAuctions();
      
      debugPrint('Reponse API - success: ${response.success}');
      debugPrint('Reponse API - count: ${response.data?.length ?? 0}');

      if (response.success && response.data != null) {
        final List<dynamic> auctionList = response.data!;
        debugPrint('${auctionList.length} encheres recues de l API');
        
        // Cache mes encheres
        await CacheService.instance.cacheMyAuctions({
          'data': auctionList,
          'success': true,
        });

        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _myAuctions = auctionList.map((item) => item as Map<String, dynamic>).toList();
          _filterAuctions();
        });
        
        debugPrint('${_myAuctions.length} encheres affichees');
      } else {
        debugPrint('Echec API: success=${response.success}, data=${response.data}');
        if (cachedMyAuctions == null && mounted) {
          setState(() => _isLoading = false);
        }
      }
    } catch (e, stackTrace) {
      debugPrint('Erreur _loadMyAuctions: $e');
      debugPrint('StackTrace: $stackTrace');
      setState(() {
        _isLoading = false;
        _error = e.toString();
      });
    }
  }
  
  Future<void> _loadMoreAuctions() async {
    if (_isLoadingMore || !_hasMore) return;
    
    setState(() {
      _isLoadingMore = true;
    });
    
    try {
      final nextPage = _currentPage + 1;
      debugPrint('Chargement page $nextPage...');
      
      // TODO: Implement pagination in API
      // For now, simulate no more data
      setState(() {
        _hasMore = false;
        _isLoadingMore = false;
      });
      
    } catch (e) {
      setState(() {
        _isLoadingMore = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        title: Text(
          l10n.text_27,
          style: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: isDarkMode ? Colors.white : Colors.black,
          ),
        ),
        leading: IconButton(
          icon: Icon(Icons.arrow_back_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        l10n.error_loading_my_auctions,
                        style: const TextStyle(color: Colors.grey),
                      ),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: _loadMyAuctions,
                        child: Text(l10n.retry),
                      ),
                    ],
                  ),
                )
              : Column(
                  children: [
                    // Barre de recherche et filtres
                    _buildSearchAndFilterBar(isDarkMode),
                    
                    // Resultats
                    Expanded(
                      child: _filteredAuctions.isEmpty
                          ? Center(
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(
                                    _searchQuery.isNotEmpty || _selectedStatus != null
                                        ? Icons.search_off
                                        : Icons.gavel_outlined,
                                    size: 64,
                                    color: Colors.grey[300],
                                  ),
                                  const SizedBox(height: 16),
                                  Text(
                                    _searchQuery.isNotEmpty || _selectedStatus != null
                                        ? 'لا توجد نتائج مطابقة'
                                        : 'لا توجد مزادات حالياً',
                                    style: TextStyle(
                                      color: Colors.grey[600],
                                      fontSize: 16,
                                      fontFamily: 'Plus Jakarta Sans',
                                    ),
                                  ),
                                  if (_searchQuery.isEmpty && _selectedStatus == null) ...[
                                    const SizedBox(height: 8),
                                    Text(
                                      'قم بإنشاء مزادك الأول!',
                                      style: TextStyle(
                                        color: Colors.grey[400],
                                        fontSize: 14,
                                        fontFamily: 'Plus Jakarta Sans',
                                      ),
                                    ),
                                    const SizedBox(height: 24),
                                    ElevatedButton.icon(
                                      onPressed: () {
                                        Navigator.of(context).push(
                                          MaterialPageRoute(
                                            builder: (context) => const CreateAdFormPage(),
                                          ),
                                        );
                                      },
                                      icon: const Icon(Icons.add),
                                      label: const Text('إنشاء مزاد جديد'),
                                    ),
                                  ],
                                ],
                              ),
                            )
                          : RefreshIndicator(
                              onRefresh: _loadMyAuctions,
                              color: const Color(0xFF0081FF),
                              child: ListView.separated(
                                controller: _scrollController,
                                padding: const EdgeInsets.all(20),
                                itemCount: _filteredAuctions.length + (_hasMore ? 1 : 0),
                                separatorBuilder: (context, index) => const SizedBox(height: 16),
                                itemBuilder: (context, index) {
                                  if (index >= _filteredAuctions.length) {
                                    return _buildLoadMoreIndicator();
                                  }
                                  final auction = _filteredAuctions[index];
                                  return _buildAuctionItem(context, auction, isDarkMode, l10n);
                                },
                              ),
                            ),
                    ),
                  ],
                ),
    );
  }

  /// Obtenir le libellé du statut en arabe
  String _getStatusLabel(String status, AppLocalizations l10n) {
    switch (status.toLowerCase()) {
      case 'active':
      case 'approved':
        return 'نشط';
      case 'pending':
        return 'قيد المراجعة';
      case 'rejected':
        return 'مرفوض';
      case 'finished':
      case 'completed':
      case 'sold':
        return 'منتهي';
      case 'cancelled':
      case 'deleted':
        return 'ملغي';
      default:
        return status;
    }
  }

  /// Obtenir la couleur du statut
  Color _getStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'active':
      case 'approved':
        return const Color(0xFF00C58D);
      case 'pending':
        return const Color(0xFFFFA500);
      case 'rejected':
      case 'cancelled':
      case 'deleted':
        return const Color(0xFFE31B23);
      case 'finished':
      case 'completed':
      case 'sold':
        return const Color(0xFF0081FF);
      default:
        return Colors.grey;
    }
  }

  /// Vérifier si l'enchère peut être modifiée
  bool _canEdit(String status) {
    final editableStatuses = ['active', 'pending', 'approved'];
    return editableStatuses.contains(status.toLowerCase());
  }

  /// Vérifier si l'enchère peut être supprimée
  bool _canDelete(String status) {
    final deletableStatuses = ['active', 'pending', 'approved', 'rejected'];
    return deletableStatuses.contains(status.toLowerCase());
  }

  /// Formater la date de fin
  String _formatEndTime(String? endTime) {
    if (endTime == null || endTime.isEmpty) return '';
    try {
      final date = DateTime.parse(endTime);
      final now = DateTime.now();
      final diff = date.difference(now);

      if (diff.isNegative) {
        return 'انتهى';
      } else if (diff.inDays > 0) {
        return '${diff.inDays} يوم متبقي';
      } else if (diff.inHours > 0) {
        return '${diff.inHours} ساعة متبقية';
      } else {
        return '${diff.inMinutes} دقيقة متبقية';
      }
    } catch (e) {
      return endTime;
    }
  }

  /// Afficher le menu d'actions
  void _showActionMenu(BuildContext context, Map<String, dynamic> auction, AppLocalizations l10n) {
    final id = auction['id']?.toString() ?? '';
    final status = auction['status']?.toString() ?? 'active';

    showModalBottomSheet(
      context: context,
      backgroundColor: Theme.of(context).brightness == Brightness.dark
          ? const Color(0xFF1D1D1D)
          : Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Barre de poignée
                Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey[300],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                const SizedBox(height: 20),

                // Titre
                Text(
                  'خيارات المزاد',
                  style: TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Theme.of(context).brightness == Brightness.dark
                        ? Colors.white
                        : Colors.black,
                  ),
                ),
                const SizedBox(height: 20),

                // Voir les détails
                _buildMenuItem(
                  context,
                  icon: Icons.visibility_outlined,
                  title: 'عرض التفاصيل',
                  onTap: () {
                    Navigator.pop(context);
                    if (id.isNotEmpty) {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => AuctionDetailsPage(auctionId: id),
                        ),
                      );
                    }
                  },
                ),

                // Voir l'historique des enchères
                _buildMenuItem(
                  context,
                  icon: Icons.history,
                  title: 'سجل المزايدات',
                  onTap: () {
                    Navigator.pop(context);
                    if (id.isNotEmpty) {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => AuctionHistoryPage(auctionId: id),
                        ),
                      );
                    }
                  },
                ),

                // Modifier (si autorisé)
                if (_canEdit(status))
                  _buildMenuItem(
                    context,
                    icon: Icons.edit_outlined,
                    title: 'تعديل المزاد',
                    color: const Color(0xFF0081FF),
                    onTap: () {
                      Navigator.pop(context);
                      _editAuction(context, auction, l10n);
                    },
                  ),

                // Supprimer (si autorisé)
                if (_canDelete(status))
                  _buildMenuItem(
                    context,
                    icon: Icons.delete_outline,
                    title: 'حذف المزاد',
                    color: const Color(0xFFE31B23),
                    onTap: () {
                      Navigator.pop(context);
                      _confirmDelete(context, id, l10n);
                    },
                  ),

                const SizedBox(height: 8),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildMenuItem(
    BuildContext context, {
    required IconData icon,
    required String title,
    required VoidCallback onTap,
    Color? color,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return ListTile(
      leading: Icon(
        icon,
        color: color ?? (isDark ? Colors.white : Colors.black87),
      ),
      title: Text(
        title,
        style: TextStyle(
          fontFamily: 'Plus Jakarta Sans',
          color: color ?? (isDark ? Colors.white : Colors.black87),
        ),
      ),
      onTap: onTap,
    );
  }

  /// Modifier une enchère
  void _editAuction(BuildContext context, Map<String, dynamic> auction, AppLocalizations l10n) {
    final id = auction['id']?.toString() ?? '';
    if (id.isEmpty) return;

    // TODO: Navigation vers la page d'édition
    // Pour l'instant, afficher un message
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('سيتم فتح صفحة التعديل قريباً'),
        backgroundColor: Color(0xFF0081FF),
      ),
    );

    // Option: Rediriger vers la page de création avec mode édition
    // Navigator.of(context).push(
    //   MaterialPageRoute(
    //     builder: (context) => CreateAdFormPage(editMode: true, auctionData: auction),
    //   ),
    // );
  }

  /// Confirmer la suppression
  void _confirmDelete(BuildContext context, String auctionId, AppLocalizations l10n) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text(
          'تأكيد الحذف',
          textAlign: TextAlign.center,
          style: TextStyle(fontFamily: 'Plus Jakarta Sans'),
        ),
        content: const Text(
          'هل أنت متأكد من حذف هذا المزاد؟ لا يمكن التراجع عن هذا الإجراء.',
          textAlign: TextAlign.center,
          style: TextStyle(fontFamily: 'Plus Jakarta Sans'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text(
              'إلغاء',
              style: TextStyle(fontFamily: 'Plus Jakarta Sans'),
            ),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deleteAuction(auctionId);
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFE31B23),
            ),
            child: const Text(
              'حذف',
              style: TextStyle(fontFamily: 'Plus Jakarta Sans'),
            ),
          ),
        ],
      ),
    );
  }

  /// Supprimer une enchère
  Future<void> _deleteAuction(String auctionId) async {
    setState(() => _isLoading = true);

    try {
      final response = await _auctionApi.deleteAuction(auctionId);

      if (response.success) {
        // Rafraîchir la liste
        await _loadMyAuctions();

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('تم حذف المزاد بنجاح'),
              backgroundColor: Color(0xFF00C58D),
            ),
          );
        }
      } else {
        setState(() => _isLoading = false);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(response.error?.message ?? 'فشل حذف المزاد'),
              backgroundColor: const Color(0xFFE31B23),
            ),
          );
        }
      }
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('خطأ: $e'),
            backgroundColor: const Color(0xFFE31B23),
          ),
        );
      }
    }
  }

  Widget _buildAuctionItem(BuildContext context, Map<String, dynamic> auction, bool isDarkMode, AppLocalizations l10n) {
    // Extraction des données API avec fallbacks
    final id = auction['id']?.toString() ?? '';

    // Récupérer le titre avec fallback intelligent
    final locale = Localizations.localeOf(context).languageCode;
    String title = '';

    switch (locale) {
      case 'ar':
        title = auction['title_ar']?.toString() ?? '';
        break;
      case 'fr':
        title = auction['title_fr']?.toString() ?? '';
        break;
      case 'en':
        title = auction['title_en']?.toString() ?? '';
        break;
    }

    if (title.isEmpty) {
      title = auction['title_ar']?.toString() ?? '';
    }

    if (title.isEmpty) {
      title = auction['title_fr']?.toString() ??
              auction['title_en']?.toString() ??
              auction['title']?.toString() ??
              'بدون عنوان';
    }

    // Les prix viennent comme strings de l'API
    final currentPriceRaw = auction['current_price'] ?? auction['current_bid'] ?? '0';
    final startPriceRaw = auction['start_price'] ?? '0';
    final currentPrice = double.tryParse(currentPriceRaw.toString())?.toStringAsFixed(0) ?? currentPriceRaw.toString();
    final startPrice = double.tryParse(startPriceRaw.toString())?.toStringAsFixed(0) ?? startPriceRaw.toString();

    final endTime = auction['end_time']?.toString() ?? auction['ends_at']?.toString() ?? '';
    final status = auction['status']?.toString() ?? 'active';
    final bidCount = auction['bidder_count'] ?? auction['bid_count'] ?? auction['bids_count'] ?? 0;
    final viewCount = auction['views'] ?? auction['view_count'] ?? 0;

     String imageUrl = '';
    bool hasImage = false;
    
    final imageUrls = auction['image_urls'];
    if (imageUrls != null && imageUrls is List && imageUrls.isNotEmpty) {
      imageUrl = imageUrls[0].toString();
      hasImage = imageUrl.isNotEmpty;
    } else if (auction['image_url'] != null && auction['image_url'].toString().isNotEmpty) {
      imageUrl = auction['image_url'].toString();
      hasImage = true;
    } else if (auction['images'] != null && auction['images'] is List && (auction['images'] as List).isNotEmpty) {
      final firstImage = (auction['images'] as List)[0];
      imageUrl = firstImage is Map ? firstImage['url']?.toString() ?? '' : firstImage.toString();
      hasImage = imageUrl.isNotEmpty;
    } else if (auction['image'] != null && auction['image'].toString().isNotEmpty) {
      imageUrl = auction['image'].toString();
      hasImage = true;
    } else if (auction['thumbnail'] != null && auction['thumbnail'].toString().isNotEmpty) {
      imageUrl = auction['thumbnail'].toString();
      hasImage = true;
    } else if (auction['main_image'] != null && auction['main_image'].toString().isNotEmpty) {
      imageUrl = auction['main_image'].toString();
      hasImage = true;
    }
    
    // Verifier le type d'image
    bool isNetworkImage = hasImage && (imageUrl.startsWith('http') || imageUrl.startsWith('https'));
    bool isDataUri = hasImage && imageUrl.startsWith('data:image');
    bool isAsset = hasImage && !isNetworkImage && !isDataUri;

    final statusColor = _getStatusColor(status);
    final statusLabel = _getStatusLabel(status, l10n);
    final timeRemaining = _formatEndTime(endTime);

    return GestureDetector(
      onTap: () {
        if (id.isNotEmpty) {
          Navigator.of(context).push(
            MaterialPageRoute(builder: (context) => AuctionDetailsPage(auctionId: id)),
          );
        }
      },
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: Colors.grey.withOpacity(0.1)),
          boxShadow: [
            BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
          ],
        ),
        child: Column(
          children: [
            Row(
              children: [
                // Image
                ClipRRect(
                  borderRadius: BorderRadius.circular(16),
                  child: hasImage
                      ? isDataUri
                          ? _buildDataUriImage(imageUrl)
                          : isNetworkImage
                              ? Image.network(
                                  imageUrl,
                                  width: 100,
                                  height: 100,
                                  fit: BoxFit.cover,
                                  loadingBuilder: (context, child, loadingProgress) {
                                    if (loadingProgress == null) return child;
                                    return Container(
                                      width: 100,
                                      height: 100,
                                      color: Colors.grey[200],
                                      child: Center(
                                        child: CircularProgressIndicator(
                                          strokeWidth: 2,
                                          value: loadingProgress.expectedTotalBytes != null
                                              ? loadingProgress.cumulativeBytesLoaded / loadingProgress.expectedTotalBytes!
                                              : null,
                                        ),
                                      ),
                                    );
                                  },
                                  errorBuilder: (c, e, s) => _buildImagePlaceholder(),
                                )
                              : isAsset
                                  ? Image.asset(
                                      imageUrl,
                                      width: 100,
                                      height: 100,
                                      fit: BoxFit.cover,
                                      errorBuilder: (c, e, s) => _buildImagePlaceholder(),
                                    )
                                  : _buildImagePlaceholder()
                      : _buildImagePlaceholder(),
                ),
                const SizedBox(width: 16),

                // Contenu
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Titre
                      Text(
                        title,
                        style: const TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 4),

                      // Catégorie et Ville
                      Row(
                        children: [
                          if (auction['category'] != null)
                            Flexible(
                              child: Text(
                                auction['category'].toString(),
                                style: TextStyle(
                                  fontFamily: 'Plus Jakarta Sans',
                                  fontSize: 11,
                                  color: Colors.grey[600],
                                ),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          if (auction['category'] != null && auction['city'] != null)
                            Text(
                              ' • ',
                              style: TextStyle(
                                fontSize: 11,
                                color: Colors.grey[400],
                              ),
                            ),
                          if (auction['city'] != null)
                            Flexible(
                              child: Text(
                                auction['city'].toString(),
                                style: TextStyle(
                                  fontFamily: 'Plus Jakarta Sans',
                                  fontSize: 11,
                                  color: Colors.grey[600],
                                ),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                        ],
                      ),
                      const SizedBox(height: 8),

                      // Prix actuel
                      Row(
                        children: [
                          Text(
                            '$currentPrice MRU',
                            style: const TextStyle(
                              fontFamily: 'Plus Jakarta Sans',
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: Color(0xFF0081FF),
                            ),
                          ),
                          if (bidCount > 0) ...[
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                              decoration: BoxDecoration(
                                color: const Color(0xFF0081FF).withOpacity(0.1),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Text(
                                '$bidCount مزايدة',
                                style: const TextStyle(
                                  fontFamily: 'Plus Jakarta Sans',
                                  fontSize: 10,
                                  color: Color(0xFF0081FF),
                                ),
                              ),
                            ),
                          ],
                        ],
                      ),
                      const SizedBox(height: 4),

                      // Prix de départ
                      Text(
                        'السعر الابتدائي: $startPrice MRU',
                        style: TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          fontSize: 12,
                          color: Colors.grey[600],
                        ),
                      ),
                      const SizedBox(height: 8),

                      // Statut et temps
                      Row(
                        children: [
                          // Badge de statut
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                            decoration: BoxDecoration(
                              color: statusColor.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Text(
                              statusLabel,
                              style: TextStyle(
                                fontFamily: 'Plus Jakarta Sans',
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                                color: statusColor,
                              ),
                            ),
                          ),
                          const SizedBox(width: 8),

                          // Temps restant
                          if (timeRemaining.isNotEmpty)
                            Flexible(
                              child: Text(
                                timeRemaining,
                                style: TextStyle(
                                  fontFamily: 'Plus Jakarta Sans',
                                  fontSize: 11,
                                  color: timeRemaining.contains('انتهى')
                                      ? const Color(0xFFE31B23)
                                      : Colors.grey[600],
                                  fontWeight: FontWeight.bold,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                        ],
                      ),
                    ],
                  ),
                ),

                // Menu d'actions
                IconButton(
                  icon: const Icon(Icons.more_vert, size: 20, color: Colors.grey),
                  onPressed: () => _showActionMenu(context, auction, l10n),
                ),
              ],
            ),

            // Statistiques en bas
            Padding(
              padding: const EdgeInsets.only(top: 12),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  // Numéro de lot
                  if (auction['lot_number'] != null)
                    _buildStatItem(Icons.confirmation_number_outlined, auction['lot_number'].toString()),

                  // Vues
                  _buildStatItem(
                    Icons.visibility_outlined,
                    '$viewCount ${viewCount == 1 ? 'مشاهدة' : 'مشاهدات'}',
                  ),

                  // Bids
                  _buildStatItem(
                    Icons.people_outline,
                    '$bidCount ${bidCount == 1 ? 'مزايد' : 'مزايدين'}',
                  ),

                  // Date de création
                  _buildStatItem(
                    Icons.access_time,
                    DateFormat('yyyy/MM/dd').format(
                      DateTime.tryParse(auction['created_at']?.toString() ?? '') ?? DateTime.now(),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatItem(IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: Colors.grey[500]),
        const SizedBox(width: 4),
        Text(
          text,
          style: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontSize: 11,
            color: Colors.grey[600],
          ),
        ),
      ],
    );
  }

  /// Barre de recherche et filtres
  Widget _buildSearchAndFilterBar(bool isDarkMode) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(20)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          // Champ de recherche
          TextField(
            controller: _searchController,
            onChanged: (value) {
              setState(() {
                _searchQuery = value;
              });
              _filterAuctions();
            },
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              color: isDarkMode ? Colors.white : Colors.black,
            ),
            decoration: InputDecoration(
              hintText: 'البحث في مزاداتي...',
              hintStyle: TextStyle(
                fontFamily: 'Plus Jakarta Sans',
                color: Colors.grey[400],
              ),
              prefixIcon: const Icon(Icons.search, color: Colors.grey),
              suffixIcon: _searchQuery.isNotEmpty
                  ? IconButton(
                      icon: const Icon(Icons.clear, color: Colors.grey),
                      onPressed: () {
                        _searchController.clear();
                        setState(() {
                          _searchQuery = '';
                        });
                        _filterAuctions();
                      },
                    )
                  : null,
              filled: true,
              fillColor: isDarkMode ? const Color(0xFF2D2D2D) : Colors.grey[100],
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: BorderSide.none,
              ),
              contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            ),
          ),
          const SizedBox(height: 12),
          
          // Filtres par statut
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                _buildStatusFilterChip('الكل', null, isDarkMode),
                const SizedBox(width: 8),
                _buildStatusFilterChip('نشط', 'active', isDarkMode),
                const SizedBox(width: 8),
                _buildStatusFilterChip('قيد المراجعة', 'pending', isDarkMode),
                const SizedBox(width: 8),
                _buildStatusFilterChip('منتهي', 'finished', isDarkMode),
                const SizedBox(width: 8),
                _buildStatusFilterChip('مرفوض', 'rejected', isDarkMode),
              ],
            ),
          ),
          
          // Compteur de resultats
          if (_searchQuery.isNotEmpty || _selectedStatus != null)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(
                '${_filteredAuctions.length} مزاد ${_filteredAuctions.length == 1 ? '' : 'ات'}',
                style: TextStyle(
                  fontFamily: 'Plus Jakarta Sans',
                  fontSize: 12,
                  color: Colors.grey[500],
                ),
              ),
            ),
        ],
      ),
    );
  }

  /// Chip de filtre par statut
  Widget _buildStatusFilterChip(String label, String? status, bool isDarkMode) {
    final isSelected = _selectedStatus == status;
    
    Color chipColor;
    if (status == null) {
      chipColor = Colors.grey;
    } else {
      chipColor = _getStatusColor(status);
    }
    
    return ChoiceChip(
      label: Text(
        label,
        style: TextStyle(
          fontFamily: 'Plus Jakarta Sans',
          fontSize: 12,
          color: isSelected ? Colors.white : isDarkMode ? Colors.white70 : Colors.black87,
        ),
      ),
      selected: isSelected,
      onSelected: (selected) {
        setState(() {
          _selectedStatus = selected ? status : null;
        });
        _filterAuctions();
      },
      selectedColor: status == null ? const Color(0xFF0081FF) : chipColor,
      backgroundColor: isDarkMode ? const Color(0xFF2D2D2D) : Colors.grey[200],
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
      ),
    );
  }

  /// Widget pour afficher une image depuis une data URI (base64)
  Widget _buildDataUriImage(String dataUri) {
    try {
      // Extraire la partie base64 de la data URI
      // Format: data:image/jpeg;base64,/9j/4AAQ...
      final commaIndex = dataUri.indexOf(',');
      if (commaIndex == -1) {
        return _buildImagePlaceholder();
      }
      
      final base64String = dataUri.substring(commaIndex + 1);
      final bytes = base64Decode(base64String);
      
      return Image.memory(
        bytes,
        width: 100,
        height: 100,
        fit: BoxFit.cover,
        errorBuilder: (c, e, s) => _buildImagePlaceholder(),
      );
    } catch (e) {
      debugPrint('Erreur decodage image base64: $e');
      return _buildImagePlaceholder();
    }
  }

  /// Widget placeholder pour les images manquantes
  Widget _buildImagePlaceholder() {
    return Container(
      width: 100,
      height: 100,
      decoration: BoxDecoration(
        color: Colors.grey[200],
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.image_outlined,
            size: 32,
            color: Colors.grey[400],
          ),
          const SizedBox(height: 4),
          Text(
            'لا توجد صورة',
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 10,
              color: Colors.grey[500],
            ),
          ),
        ],
      ),
    );
  }

  /// Indicateur de chargement pour "charger plus"
  Widget _buildLoadMoreIndicator() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Center(
        child: _isLoadingMore
            ? const CircularProgressIndicator()
            : TextButton.icon(
                onPressed: _hasMore ? _loadMoreAuctions : null,
                icon: const Icon(Icons.expand_more),
                label: Text(
                  _hasMore ? 'عرض المزيد' : 'لا يوجد مزادات اضافية',
                  style: const TextStyle(fontFamily: 'Plus Jakarta Sans'),
                ),
              ),
      ),
    );
  }
}
