<?php

namespace App\Telegram;

use SergiX44\Nutgram\Conversations\Conversation;
use SergiX44\Nutgram\Nutgram;
use SergiX44\Nutgram\Telegram\Properties\ParseMode;
use SergiX44\Nutgram\Telegram\Types\Keyboard\InlineKeyboardButton;
use SergiX44\Nutgram\Telegram\Types\Keyboard\InlineKeyboardMarkup;

class RulesConversation extends Conversation {
    public function start(Nutgram $bot) {
        $bot->sendMessage("*Правила*\n\n1\. При первом контакте с ботом на ваш счет закидывается 50 монет\n2\. За каждый верно отвеченный вопрос со счета снимается 1 монета\n3\. Ответом является одно слово на русском языке в именительном падеже единственного числа, если в вопросе не указано иное\. \n4\. Регистр букв в ответе не имеет значения\. \n5\. За каждое нажатие кнопки Сдаюсь со счета снимается 1 монета\.\n6\. Счет привязан не к пользователю, а к чату\.\n7\. Бот поставляется \"как есть\"\. Администрация не несет ответственности за любые негативные последствия, прямо или косвенно вызванные использованием бота\.", parse_mode: ParseMode::MARKDOWN);
    }
}
