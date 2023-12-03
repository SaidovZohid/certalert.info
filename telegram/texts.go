package telegram

// languages
const (
	langUz  = "uz"
	langRu  = "ru"
	langEng = "eng"
)

// steps
const (
	stepLang = "lang"
)

// keyboards's texts
const (
	DisplayLangUz  = "🇺🇿 O'zbekcha"
	DisplayLangRu  = "🇷🇺 Русский"
	DisplayLangEng = "🇬🇧 English"
)

// messages
var (
	uzNotValidUserID             = "🙅‍♂️ Kechirasiz, siz kiritgan foydalanuvchi identifikatori yaroqsiz. Davom etish uchun to'g'ri foydalanuvchi ID isini taqdim eting. 🔄"
	uzNotFoundUserID             = "🔍 Kechirasiz, foydalanuvchi topilmadi. Qayta urinib ko'ring. 🔄"
	uzInternalErrorMsg           = "🛑 Uzr, kutilmagan xato yuz berdi. Iltimos, keyinroq qayta urinib ko'ring yoki yordam uchun qo'llab-quvvatlash xizmatiga murojaat qiling. Agar bu muammo davom etsa, @zohid_0212 ga qo'llab-quvvatlash uchun murojaat qiling. 🚀"
	uzAlreadyLinkedToThisAccount = "🔗 Sizning hisobingiz allaqachon bog'langan. 🔗"

	ruNotValidUserID             = "🙅‍♂️ Извините, введенный вами идентификатор пользователя неверен. Пожалуйста, укажите правильный идентификатор пользователя, чтобы продолжить. 🔄"
	ruNotFoundUserID             = "🔍 К сожалению, пользователь не был найден. Пожалуйста, повторите попытку с действительным идентификатором пользователя. 🔄"
	ruInternalErrorMsg           = "🛑 Приносим извинения, произошла непредвиденная ошибка. Пожалуйста, повторите попытку позже или обратитесь за помощью в службу поддержки. Если эта проблема не устранена, пожалуйста, обратитесь в службу поддержки по адресу @zohid_0212. 🚀"
	ruAlreadyLinkedToThisAccount = "🔗 Ваша учетная запись уже привязана. Чтобы изучить дополнительные команды, пожалуйста, используйте другие доступные опции или введите /help для получения помощи. 🚀"

	engNotValidUserID             = "🙅‍♂️ Sorry, the user ID you've entered is invalid. Please provide a correct user ID to proceed. 🔄"
	engNotFoundUserID             = "🔍 Sorry, the user was not found. Please try again with a valid user ID. 🔄"
	engInternalErrorMsg           = "🛑 Apologies, an unexpected error occurred. Please try again later or contact support for assistance. If this issue persists, please reach out to support at @zohid_0212. 🚀"
	engAlreadyLinkedToThisAccount = "🔗 Your account is already linked. To explore further commands, please use other available options or type /help for assistance. 🚀"
)
