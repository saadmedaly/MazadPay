package services

// NotificationLocalization holds translations for notification messages
type NotificationLocalization struct {
	Title string
	Body  string
}

// NotificationLocalizations maps notification types to language-specific messages
var NotificationLocalizations = map[string]map[string]NotificationLocalization{
	"auction_pending": {
		"ar": {
			Title: "مزاد جديد في الانتظار",
			Body:  "{userName} أنشأ مزاد: {auctionTitle}",
		},
		"fr": {
			Title: "Nouvelle enchère en attente",
			Body:  "{userName} a créé une enchère: {auctionTitle}",
		},
		"en": {
			Title: "New auction pending",
			Body:  "{userName} created an auction: {auctionTitle}",
		},
	},
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
	"auction_ending_soon": {
		"ar": {
			Title: "⚡ الفرصة الأخيرة!",
			Body:  "\"{auctionTitle}\" ينتهي في 5 دقائق",
		},
		"fr": {
			Title: "⚡ Dernière chance !",
			Body:  "\"{auctionTitle}\" se termine dans 5 minutes",
		},
		"en": {
			Title: "⚡ Last chance!",
			Body:  "\"{auctionTitle}\" ends in 5 minutes",
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
			Title: "🚨 مزاد مُبلّغ عنه",
			Body:  "إبلاغ من {reporter} على \"{auctionTitle}\"",
		},
		"fr": {
			Title: "🚨 Enchère signalée",
			Body:  "Signalement de {reporter} sur \"{auctionTitle}\"",
		},
		"en": {
			Title: "🚨 Auction reported",
			Body:  "Report from {reporter} on \"{auctionTitle}\"",
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
