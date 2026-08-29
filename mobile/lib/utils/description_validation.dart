/// Pure helpers for the ad-description length constraint (10-5000 chars,
/// mirroring the backend requirement in `create_ad_form_page.dart`).
/// Extracted so the limits can be unit-tested without pumping the whole
/// form widget.
library;

const int kDescriptionMinLength = 10;
const int kDescriptionMaxLength = 5000;

bool isDescriptionTooShort(String description) =>
    description.trim().length < kDescriptionMinLength;

bool isDescriptionTooLong(String description) =>
    description.length > kDescriptionMaxLength;

/// True when [description] satisfies both bounds — used when a submission
/// is not a draft (drafts may have an incomplete description for now).
bool isDescriptionValid(String description) {
  final trimmed = description.trim();
  return trimmed.length >= kDescriptionMinLength &&
      description.length <= kDescriptionMaxLength;
}
