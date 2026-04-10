import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:video_player/video_player.dart';

class HowToBidPage extends StatefulWidget {
  const HowToBidPage({super.key});

  @override
  State<HowToBidPage> createState() => _HowToBidPageState();
}

class _HowToBidPageState extends State<HowToBidPage> {
  int _selectedVideoIndex = 0;
  late VideoPlayerController _videoController;
  bool _isInitialized = false;

  final List<Map<String, String>> _tutorials = [
    {
      'title': 'كيف يمكنني دفع من من أي تطبيق بنكي',
      'thumbnail': 'assets/mr.png', // Small flag or Bankily logo placeholder
      'videoUrl': 'assets/MezadPay.mp4',
    },
    {
      'title': 'كيفية المزايدة على السيارات',
      'thumbnail': '',
      'videoUrl': '',
    },
    {
      'title': 'طريقة استلام سيارتك بعد الفوز',
      'thumbnail': '',
      'videoUrl': '',
    },
    {
      'title': 'شرح نظام العمولات والشحن',
      'thumbnail': '',
      'videoUrl': '',
    },
    {
      'title': 'الأسئلة الشائعة',
      'thumbnail': '',
      'videoUrl': '',
    },
  ];

  @override
  void initState() {
    super.initState();
    _initializeVideo();
  }

  void _initializeVideo() {
    _isInitialized = false;
    final videoPath = _tutorials[_selectedVideoIndex]['videoUrl'];
    if (videoPath != null && videoPath.isNotEmpty) {
      _videoController = VideoPlayerController.asset(videoPath)
        ..initialize().then((_) {
          setState(() {
            _isInitialized = true;
          });
        });
    }
  }

  @override
  void dispose() {
    _videoController.dispose();
    super.dispose();
  }

  void _onTutorialSelected(int index) {
    if (_selectedVideoIndex == index) return;
    
    setState(() {
      _selectedVideoIndex = index;
    });
    
    _videoController.dispose();
    _initializeVideo();
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          leading: Padding(
            padding: const EdgeInsets.all(8.0),
            child: Container(
              decoration: BoxDecoration(
                color: isDarkMode ? Colors.grey[900] : Colors.grey[100],
                shape: BoxShape.circle,
              ),
              child: IconButton(
                icon: Icon(
                  Icons.arrow_back_ios,
                  size: 16,
                  color: isDarkMode ? Colors.white : Colors.black,
                ),
                onPressed: () => Navigator.of(context).pop(),
              ),
            ),
          ),
          title: Text(
            'كيفية مزايدة والشحن',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
          ),
          centerTitle: true,
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildSectionHeader(context, 'فيديوهات تعلمية'),
              const SizedBox(height: 16),
              
              // Video Player Section
              _buildVideoPlayerSection(context),
              
              const SizedBox(height: 16),
              const Divider(height: 32, thickness: 1, color: Colors.black12),
              
              // List of tutorials
              ...List.generate(_tutorials.length, (index) {
                return _buildTutorialItem(
                  context,
                  index: index,
                  title: _tutorials[index]['title']!,
                  thumbnail: _tutorials[index]['thumbnail']!,
                  isSelected: _selectedVideoIndex == index,
                  onTap: () => _onTutorialSelected(index),
                );
              }),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSectionHeader(BuildContext context, String title) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Row(
      mainAxisAlignment: MainAxisAlignment.end,
      children: [
        Text(
          title,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 16,
            fontWeight: FontWeight.w700,
            color: isDarkMode ? Colors.white : Colors.black,
          ),
        ),
        const SizedBox(width: 8),
        const Icon(Icons.play_circle_fill, color: Colors.red, size: 28),
      ],
    );
  }

  Widget _buildVideoPlayerSection(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return Container(
      width: double.infinity,
      height: 220,
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.08),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Stack(
          alignment: Alignment.center,
          children: [
            // Video Background or Thumbnail Placeholder
            if (_isInitialized)
              AspectRatio(
                aspectRatio: _videoController.value.aspectRatio,
                child: VideoPlayer(_videoController),
              )
            else
              Container(
                color: isDarkMode ? Colors.grey[900] : Colors.grey[200],
                child: Center(
                  child: Image.asset(
                    'assets/logo.png', // Main branding
                    width: 140,
                    opacity: const AlwaysStoppedAnimation(0.2),
                    errorBuilder: (context, error, stackTrace) => const Icon(Icons.play_circle_outline, size: 64, color: Colors.grey),
                  ),
                ),
              ),

            // Video Overlay UI
            if (!_isInitialized || !_videoController.value.isPlaying)
              Container(
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.15),
                ),
              ),

            // Skip Buttons 
            Positioned(
              left: 40,
              child: _buildSkipButton(Icons.replay_10),
            ),
            Positioned(
              right: 40,
              child: _buildSkipButton(Icons.forward_10),
            ),

            // Play Button
            Center(
              child: InkWell(
                onTap: () {
                  if (_isInitialized) {
                    setState(() {
                      _videoController.value.isPlaying ? _videoController.pause() : _videoController.play();
                    });
                  }
                },
                child: Container(
                  height: 60,
                  width: 60,
                  decoration: BoxDecoration(
                    color: Colors.black.withOpacity(0.4),
                    shape: BoxShape.circle,
                    border: Border.all(color: Colors.white, width: 2),
                  ),
                  child: Icon(
                    _isInitialized && _videoController.value.isPlaying ? Icons.pause : Icons.play_arrow,
                    color: Colors.white, 
                    size: 36
                  ),
                ),
              ),
            ),

            // Branding Text/Logo (from screenshot)
            Positioned(
              top: 40,
              child: _isInitialized ? const SizedBox() : Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text('azad', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 32)),
                  Text('Pay', style: TextStyle(color: Theme.of(context).primaryColor, fontWeight: FontWeight.bold, fontSize: 32)),
                ],
              ),
            ),

            // Progress Bar (from screenshot)
            Positioned(
              bottom: 0,
              left: 0,
              right: 0,
              child: Container(
                height: 48,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                decoration: BoxDecoration(
                  color: const Color(0xFF0081FF).withOpacity(0.9),
                ),
                child: Row(
                  children: [
                    const Text('0:00', style: TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold)),
                    Expanded(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        child: Container(
                          height: 4,
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.3),
                            borderRadius: BorderRadius.circular(2),
                          ),
                          child: FractionallySizedBox(
                            alignment: Alignment.centerRight,
                            widthFactor: 0.35, // Mocked progress
                            child: Container(
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(2),
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                    const Icon(Icons.more_horiz, color: Colors.white, size: 24),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSkipButton(IconData icon) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.3),
        shape: BoxShape.circle,
      ),
      child: Icon(icon, color: Colors.white, size: 24),
    );
  }

  Widget _buildTutorialItem(BuildContext context, {
    required int index,
    required String title, 
    required String thumbnail, 
    required bool isSelected,
    required VoidCallback onTap 
  }) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: isSelected ? Border.all(color: const Color(0xFF0081FF).withOpacity(0.5), width: 2) : null,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(isSelected ? 0.08 : 0.03),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          children: [
            // Left Chevron
            Icon(
              Icons.chevron_left,
              size: 24,
              color: isDarkMode ? Colors.white54 : Colors.black26,
            ),
            const SizedBox(width: 8),
            // Title text
            Expanded(
              child: Text(
                title,
                textAlign: TextAlign.center,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: isDarkMode ? Colors.white : Colors.black,
                ),
              ),
            ),
            const SizedBox(width: 12),
            // Thumbnail on the Right (RTL)
            Container(
              width: 65,
              height: 45,
              decoration: BoxDecoration(
                color: isDarkMode ? Colors.grey[800] : const Color(0xFFF2F4F7),
                borderRadius: BorderRadius.circular(12),
              ),
              child: thumbnail.isNotEmpty
                ? ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: Image.asset(
                    thumbnail,
                    fit: BoxFit.contain,
                    errorBuilder: (context, error, stackTrace) => const Icon(Icons.image, color: Colors.grey),
                  ),
                )
                : Icon(
                    title == 'الأسئلة الشائعة'
                        ? Icons.help_outline
                        : Icons.play_circle_outline,
                    color: Colors.grey,
                  ),
            ),
          ],
        ),
      ),
    );
  }
}
