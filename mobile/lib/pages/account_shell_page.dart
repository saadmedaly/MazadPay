import 'package:flutter/material.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';


/// A scaffold that keeps the global drawer and bottom navigation bar
/// while showing the AccountPage as its body.
class AccountShellPage extends StatelessWidget {
  const AccountShellPage({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      drawer: const SideMenuDrawer(),
      // You may also want to provide an appBar if your design requires it.
      body: const AccountPage(),
      // Highlight the Account tab (index may vary in your BottomNavBar implementation).
      bottomNavigationBar: const BottomAppBar(),
    );
  }
}
