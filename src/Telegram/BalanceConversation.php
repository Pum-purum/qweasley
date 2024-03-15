<?php

namespace App\Telegram;

use App\Entity\Chat;
use SergiX44\Nutgram\Conversations\Conversation;
use SergiX44\Nutgram\Nutgram;
use SergiX44\Nutgram\Telegram\Properties\ParseMode;

class BalanceConversation extends Conversation {
    public function start(Nutgram $bot) {
        $chat = em()->getRepository(Chat::class)->findOneBy(['telegramId' => $bot->chatId()]);
        $bot->sendMessage("*Ваш баланс: " . $chat->getBalance() . " монет\\.*\n\nПополнить баланс вы можете, предложив свой вопрос через соответствующую команду меню\\. В случае, если вопрос пройдет модерацию, он будет опубликован в боте и ваш счет будет пополнен на 10 монет\\. Если вы готовы приобрести монеты за деньги по курсу 1 монета \\= 5 рублей, свяжитесь с администрацией через команду \\/feedback", parse_mode: ParseMode::MARKDOWN);
    }

    public function beforeStep(Nutgram $bot): void {
        em()->clear();
        $this->refreshOnDeserialize();

        parent::beforeStep($bot);
    }
}
