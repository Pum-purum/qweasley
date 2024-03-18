<?php

namespace App\Telegram;

use App\Entity\Chat;
use App\Entity\FeedBack;
use SergiX44\Nutgram\Conversations\Conversation;
use SergiX44\Nutgram\Nutgram;

class FeedBackConversation extends Conversation {

    public string $action = '';

    public function start(Nutgram $bot) {
        $chat = em()->getRepository(Chat::class)->findOneBy(['telegramId' => $bot->chatId()]);
        if (null === $chat) {
            $chat = new Chat();
            $chat->setTelegramId($bot->chatId());

            em()->persist($chat);
            em()->flush();
        }

        $bot->sendMessage('Если у вас есть вопросы, предложения или жалобы, напишите их следующим сообщением. Мы обязательно их увидим.');

        $this->next('waitResponse');
    }

    public function waitResponse(Nutgram $bot) {
        if (null !== $bot->callbackQuery() || null === $bot->message() || mb_strlen((string)$bot->message()->text) < 3) {
            $bot->sendMessage('К сожалению, это некорректное сообщение.');
            $this->end();

            return;
        }

        $bot->sendMessage('Ваше сообщение принято! Спасибо.');

        $chat = em()->getRepository(Chat::class)->findOneBy(['telegramId' => $bot->chatId()]);
        $feedback = new FeedBack();
        $feedback->setChat($chat);
        $feedback->setText($bot->message()->text);

        em()->persist($feedback);
        em()->flush();

        $bot->sendMessage('Новое сообщение в форме обратной связи', $_ENV['ADMIN_CHAT_ID']);

        $this->end();
    }
}
