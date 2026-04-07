import 'package:flutter_riverpod/flutter_riverpod.dart';

class LoginState {
  final bool isLoading;
  final String? error;

  LoginState({this.isLoading = false, this.error});

  LoginState copyWith({bool? isLoading, String? error}) {
    return LoginState(
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
    );
  }
}

class LoginController extends StateNotifier<LoginState> {
  LoginController() : super(LoginState());

  Future<bool> login(String phone, String password) async {
    state = state.copyWith(isLoading: true, error: null);
    
    try {
      // Simulation d'appel API (à remplacer par un vrai appel auth plus tard)
      await Future.delayed(const Duration(seconds: 2));
      
      // Ici, on accepte n'importe quoi pour la démo, 
      // ou on met des identifiants de test.
      if (phone.length >= 8 && password.length == 4) {
        state = state.copyWith(isLoading: false);
        return true;
      } else {
        state = state.copyWith(isLoading: false, error: "يرجى التحقق من البيانات");
        return false;
      }
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }
}

final loginControllerProvider = StateNotifierProvider<LoginController, LoginState>((ref) {
  return LoginController();
});
