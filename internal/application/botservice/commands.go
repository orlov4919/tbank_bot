package botservice

const (
	Start   = "/start"   // Регистрация пользователя
	Help    = "/help"    // Вывод списка доступных команд.
	Track   = "/track"   // Начать отслеживание ссылки
	Untrack = "/untrack" //  Прекратить отслеживание ссылки.
	List    = "/list"    // Показать список отслеживаемых ссылок (cписок ссылок, полученных при /track)
)

var commandsDescription = [][2]string{
	{Start, "начало общения с ботом"},
	{Help, "вывод всех команд"},
	{Track, "начать отслеживать ссылку"},
	{Untrack, "перестать отслеживать ссылку"},
	{List, "список сохраненных ссылок"},
}
