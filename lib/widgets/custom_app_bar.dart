import 'package:flutter/material.dart';

class CustomAppBar extends StatelessWidget implements PreferredSizeWidget {
  final String? title;
  final bool showMenu;
  final bool showBack;
  final List<Widget>? actions;
  final VoidCallback? onBackPress;

  const CustomAppBar({
    super.key,
    this.title,
    this.showMenu = true,
    this.showBack = false,
    this.actions,
    this.onBackPress,
  });

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return AppBar(
      backgroundColor: Colors.transparent,
      elevation: 0,
      toolbarHeight: 70,
      automaticallyImplyLeading: false,
      title: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Leading section
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (showMenu) ...[
                Builder(
                  builder: (context) => IconButton(
                    icon: Icon(
                      Icons.menu,
                      color: isDarkMode ? Colors.white : Colors.black,
                      size: 28,
                    ),
                    onPressed: () => Scaffold.of(context).openDrawer(),
                  ),
                ),
                IconButton(
                  icon: Icon(
                    Icons.notifications_outlined,
                    color: isDarkMode ? Colors.white : Colors.black,
                    size: 28,
                  ),
                  onPressed: () {},
                ),
              ],
              if (showBack)
                IconButton(
                  icon: Icon(
                    Icons.adaptive.arrow_back,
                    color: isDarkMode ? Colors.white : Colors.black,
                    size: 24,
                  ),
                  onPressed: onBackPress ?? () => Navigator.maybePop(context),
                ),
              if (title != null) ...[
                const SizedBox(width: 8),
                Text(
                  title!,
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 18,
                    color: isDarkMode ? Colors.white : Colors.black,
                  ),
                ),
              ],
            ],
          ),
          
          // Center section (Logo)
          if (title == null)
            Image.asset(
              'logo.png',
              height: 35,
              errorBuilder: (c, e, s) => const Text(
                'MazadPay',
                style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold),
              ),
            ),
            
          // Trailing section
          Row(
            mainAxisSize: MainAxisSize.min,
            children: actions ?? [],
          ),
        ],
      ),
    );
  }

  @override
  Size get preferredSize => const Size.fromHeight(70);
}
