/// Client feedback: Bug H. Several auction-card/list surfaces defaulted a
/// missing image to a bundled stock photo (assets/corolla.png), which
/// rendered successfully and looked exactly like a real auction photo --
/// e.g. auction "حرامي" (whose real R2 image object was deleted, see Bug G)
/// showed a Toyota Corolla photo on Home/Live Auction and Auction History
/// cards, misleading users into thinking that was the auction's real image.
///
/// This is the single, pure decision: given an auction's list of image
/// URLs, what (if any) real URL should be displayed? An empty result means
/// "no real image -- render the neutral missing-image placeholder", never
/// a fake stock asset. Screens that render network vs. local-asset images
/// differently still branch on the URL's own scheme; this only decides
/// whether a real URL exists at all.
String? resolveAuctionImageUrl(List<String> imageUrls) {
  for (final url in imageUrls) {
    if (url.trim().isNotEmpty) return url;
  }
  return null;
}
