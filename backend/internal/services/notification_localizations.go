package services

// NotificationLocalization holds translations for notification messages
type NotificationLocalization struct {
	Title string
	Body  string
}

// NotificationLocalizations maps notification types to language-specific messages
var NotificationLocalizations = map[string]map[string]NotificationLocalization{
	"auction_approved": {
		"ar": {
			Title: "تمت الموافقة على المزاد!",
			Body:  "مزادك \"{auctionTitle}\" أصبح متاحًا الآن",
		},
		"fr": {
			Title: "Enchère approuvée !",
			Body:  "Votre enchère \"{auctionTitle}\" est maintenant en ligne",
		},
		"en": {
			Title: "Auction approved!",
			Body:  "Your auction \"{auctionTitle}\" is now live",
		},
	},
	"auction_rejected": {
		"ar": {
			Title: "تم رفض المزاد",
			Body:  "السبب: {reason}",
		},
		"fr": {
			Title: "Enchère refusée",
			Body:  "Raison: {reason}",
		},
		"en": {
			Title: "Auction rejected",
			Body:  "Reason: {reason}",
		},
	},
	"auction_won": {
		"ar": {
			Title: "تهانينا! لقد فزت",
			Body:  "مزاد \"{auctionTitle}\" - {finalPrice} MRU",
		},
		"fr": {
			Title: "Félicitations ! Vous avez gagné",
			Body:  "Enchère \"{auctionTitle}\" - {finalPrice} MRU",
		},
		"en": {
			Title: "Congratulations! You won",
			Body:  "Auction \"{auctionTitle}\" - {finalPrice} MRU",
		},
	},
	"auction_ended": {
		"ar": {
			Title: "انتهى المزاد",
			Body:  "تم بيع \"{auctionTitle}\" بـ {finalPrice} MRU",
		},
		"fr": {
			Title: "Enchère terminée",
			Body:  "\"{auctionTitle}\" vendu pour {finalPrice} MRU",
		},
		"en": {
			Title: "Auction ended",
			Body:  "\"{auctionTitle}\" sold for {finalPrice} MRU",
		},
	},
	"payment_received": {
		"ar": {
			Title: "💰 تم استلام الدفع",
			Body:  "{amount} MRU لمزاد \"{auctionTitle}\"",
		},
		"fr": {
			Title: "💰 Paiement reçu",
			Body:  "{amount} MRU pour \"{auctionTitle}\"",
		},
		"en": {
			Title: "💰 Payment received",
			Body:  "{amount} MRU for \"{auctionTitle}\"",
		},
	},
	"new_message": {
		"ar": {
			Title: "رسالة جديدة من {senderName}",
			Body:  "{messagePreview}...",
		},
		"fr": {
			Title: "Nouveau message de {senderName}",
			Body:  "{messagePreview}...",
		},
		"en": {
			Title: "New message from {senderName}",
			Body:  "{messagePreview}...",
		},
	},
	"auction_reported": {
		"ar": {
			Title: "🚨 بلاغ جديد",
			Body:  "تم الإبلاغ عن \"{auctionTitle}\". السبب: {reason}",
		},
		"fr": {
			Title: "🚨 Nouveau signalement",
			Body:  "Signalement de \"{auctionTitle}\". Raison: {reason}",
		},
		"en": {
			Title: "🚨 New report",
			Body:  "Report on \"{auctionTitle}\". Reason: {reason}",
		},
	},
	"banner_approved": {
		"ar": {
			Title: "تم قبول طلب الإعلان",
			Body:  "تم قبول طلبك لإضافة الإعلان {bannerTitle}",
		},
		"fr": {
			Title: "Publicité approuvée",
			Body:  "Votre demande de publicité {bannerTitle} a été acceptée",
		},
		"en": {
			Title: "Banner approved",
			Body:  "Your banner request {bannerTitle} has been approved",
		},
	},
	"banner_rejected": {
		"ar": {
			Title: "تم رفض طلب الإعلان",
			Body:  "تم رفض طلبك لإضافة الإعلان {bannerTitle}",
		},
		"fr": {
			Title: "Publicité refusée",
			Body:  "Votre demande de publicité {bannerTitle} a été refusée",
		},
		"en": {
			Title: "Banner rejected",
			Body:  "Your banner request {bannerTitle} has been rejected",
		},
	},
	"bid_outbid": {
		"ar": {
			Title: "تم تجاوز مزايدتك!",
			Body:  "تم تجاوز مزايدتك في مزاد \"{auctionTitle}\" بسعر {newPrice} MRU",
		},
		"fr": {
			Title: "Enchère dépassée !",
			Body:  "Votre enchère sur \"{auctionTitle}\" a été dépassée à {newPrice} MRU",
		},
		"en": {
			Title: "You've been outbid!",
			Body:  "Your bid on \"{auctionTitle}\" has been outbid at {newPrice} MRU",
		},
	},
	"auction_lost": {
		"ar": {
			Title: "لم تفز بالمزاد",
			Body:  "للأسف، لم تفز بمزاد \"{auctionTitle}\". السعر النهائي: {finalPrice} MRU",
		},
		"fr": {
			Title: "Enchère perdue",
			Body:  "Désolé, vous n'avez pas remporté l'enchère \"{auctionTitle}\". Prix final: {finalPrice} MRU",
		},
		"en": {
			Title: "Auction lost",
			Body:  "Sorry, you didn't win auction \"{auctionTitle}\". Final price: {finalPrice} MRU",
		},
	},
	"banner_request": {
		"ar": {
			Title: "💰 طلب إعلان جديد",
			Body:  "طلب جديد لإعلان: {bannerTitle}",
		},
		"fr": {
			Title: "💰 Nouvelle demande de publicité",
			Body:  "Nouvelle demande de publicité: {bannerTitle}",
		},
		"en": {
			Title: "💰 New banner request",
			Body:  "New banner request: {bannerTitle}",
		},
	},
	"deposit_confirmed": {
		"ar": {
			Title: "✅ تم تأكيد الإيداع",
			Body:  "تم تأكيد إيداع {amount} MRU في محفظتك",
		},
		"fr": {
			Title: "✅ Dépôt confirmé",
			Body:  "Votre dépôt de {amount} MRU a été confirmé",
		},
		"en": {
			Title: "✅ Deposit confirmed",
			Body:  "Your deposit of {amount} MRU has been confirmed",
		},
	},
	"deposit_rejected": {
		"ar": {
			Title: "❌ تم رفض الإيداع",
			Body:  "تم رفض إيداع {amount} MRU. السبب: {reason}",
		},
		"fr": {
			Title: "❌ Dépôt refusé",
			Body:  "Votre dépôt de {amount} MRU a été refusé. Raison: {reason}",
		},
		"en": {
			Title: "❌ Deposit rejected",
			Body:  "Your deposit of {amount} MRU was rejected. Reason: {reason}",
		},
	},
}

// GetLocalizedNotification retrieves a localized notification by type and language
func GetLocalizedNotification(notificationType, language string, params map[string]string) (title, body string) {
	// Default to Arabic if language not found
	locales, ok := NotificationLocalizations[notificationType]
	if !ok {
		return "", ""
	}

	localization, ok := locales[language]
	if !ok {
		// Fallback to English if language not available
		localization = locales["en"]
		// If English not available, use Arabic
		if localization.Title == "" {
			localization = locales["ar"]
		}
	}

	title = localization.Title
	body = localization.Body

	// Replace parameters
	for key, value := range params {
		placeholder := "{" + key + "}"
		title = replaceAll(title, placeholder, value)
		body = replaceAll(body, placeholder, value)
	}

	return title, body
}

// Simple string replacement helper
func replaceAll(s, old, new string) string {
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			s = s[:i] + new + s[i+len(old):]
			i += len(new) - 1
		}
	}
	return s
}
