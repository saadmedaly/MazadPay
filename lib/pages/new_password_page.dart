import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'login_page.dart';

class NewPasswordPage extends StatefulWidget {
  const NewPasswordPage({super.key});

  @override
  State<NewPasswordPage> createState() => _NewPasswordPageState();
}

class _NewPasswordPageState extends State<NewPasswordPage> {
  // Focus nodes for navigating between boxes
  final List<FocusNode> _passNodes = List.generate(4, (index) => FocusNode());
  final List<FocusNode> _confirmNodes = List.generate(4, (index) => FocusNode());

  // Controllers
  final List<TextEditingController> _passControllers =
      List.generate(4, (index) => TextEditingController());
  final List<TextEditingController> _confirmControllers =
      List.generate(4, (index) => TextEditingController());

  @override
  void dispose() {
    for (var node in _passNodes) {
      node.dispose();
    }
    for (var node in _confirmNodes) {
      node.dispose();
    }
    for (var controller in _passControllers) {
      controller.dispose();
    }
    for (var controller in _confirmControllers) {
      controller.dispose();
    }
    super.dispose();
  }

  void _nextField({required String value, required FocusNode focusNode}) {
    if (value.length == 1) {
      focusNode.requestFocus();
    }
  }

  Widget _buildBoxField(
      {required TextEditingController controller,
      required FocusNode focusNode,
      FocusNode? nextFocusNode}) {
    return Container(
      width: 50,
      height: 60,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade400, width: 1.5),
      ),
      child: Center(
        child: TextField(
          controller: controller,
          focusNode: focusNode,
          keyboardType: TextInputType.number,
          textAlign: TextAlign.center,
          style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
          inputFormatters: [
            LengthLimitingTextInputFormatter(1),
            FilteringTextInputFormatter.digitsOnly,
          ],
          onChanged: (value) {
            if (value.isNotEmpty && nextFocusNode != null) {
              _nextField(value: value, focusNode: nextFocusNode);
            }
          },
          decoration: const InputDecoration(
            border: InputBorder.none,
            counterText: "",
          ),
        ),
      ),
    );
  }

  Widget _buildOtpRow(
      List<TextEditingController> controllers, List<FocusNode> nodes) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(4, (index) {
        return Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4.0),
          child: _buildBoxField(
            controller: controllers[index],
            focusNode: nodes[index],
            nextFocusNode: index < 3 ? nodes[index + 1] : null,
          ),
        );
      }),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: Colors.white,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          leading: Padding(
            padding: const EdgeInsets.only(right: 16.0),
            child: Center(
              child: InkWell(
                onTap: () => Navigator.pop(context),
                borderRadius: BorderRadius.circular(20),
                child: Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: Colors.grey[100],
                  ),
                  child: const Icon(
                    Icons.chevron_right,
                    color: Colors.grey,
                  ),
                ),
              ),
            ),
          ),
          title: const Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'هل تحتاج مساعدة؟',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: Colors.grey,
                ),
              ),
              Text(
                'تواصل مع الدعم',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: Colors.black,
                ),
              ),
            ],
          ),
          actions: [
            Padding(
              padding: const EdgeInsets.only(left: 16.0),
              child: Center(
                child: Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: Colors.grey[100],
                  ),
                  child: const Icon(
                    Icons.smart_toy_outlined,
                    size: 20,
                    color: Colors.grey,
                  ),
                ),
              ),
            ),
          ],
        ),
        body: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                // Illustration Placeholder
                _buildIconIllustration(context),

                const SizedBox(height: 32),

                const Text(
                  'كلمة المرور الجديدة',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.black,
                  ),
                ),
                const SizedBox(height: 16),
                Directionality(
                  textDirection: TextDirection.ltr,
                  child: _buildOtpRow(_passControllers, _passNodes),
                ),

                const SizedBox(height: 32),

                const Text(
                  'قم بتأكيد كلمة المرور الجديدة',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.black,
                  ),
                ),
                const SizedBox(height: 16),
                Directionality(
                  textDirection: TextDirection.ltr,
                  child: _buildOtpRow(_confirmControllers, _confirmNodes),
                ),

                const SizedBox(height: 48),

                SizedBox(
                  width: double.infinity,
                  height: 50,
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.pushAndRemoveUntil(
                        context,
                        MaterialPageRoute(
                          builder: (context) => const LoginPage(showSuccessDialog: true),
                        ),
                        (route) => false,
                      );
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0081FF),
                      elevation: 0,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                    ),
                    child: const Text(
                      'التالي',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildIconIllustration(BuildContext context) {
    return SizedBox(
      height: 180,
      width: double.infinity,
      child: Stack(
        alignment: Alignment.center,
        children: [
          Icon(Icons.refresh, size: 160, color: Colors.blue.withOpacity(0.1)),
          const Icon(Icons.lock_reset_rounded, size: 80, color: Color(0xFF0081FF)),
          Positioned(
            top: 40,
            right: MediaQuery.of(context).size.width * 0.3,
            child: Icon(Icons.key_rounded, size: 40, color: Colors.amber.withOpacity(0.6)),
          ),
        ],
      ),
    );
  }
}
