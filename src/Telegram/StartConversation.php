<?php

namespace App\Telegram;

use App\Entity\Chat;
use App\Entity\Picture;
use App\Entity\Question;
use App\Entity\Reaction;
use SergiX44\Nutgram\Conversations\Conversation;
use SergiX44\Nutgram\Nutgram;
use SergiX44\Nutgram\Telegram\Properties\ParseMode;
use SergiX44\Nutgram\Telegram\Types\Internal\InputFile;
use SergiX44\Nutgram\Telegram\Types\Keyboard\InlineKeyboardButton;
use SergiX44\Nutgram\Telegram\Types\Keyboard\InlineKeyboardMarkup;
use SergiX44\Nutgram\Telegram\Types\Message\Message;

class StartConversation extends Conversation {

    public string $action = '';

    protected int $questionId;

    private function chat(): Chat {
        $chat = em()->getRepository(Chat::class)->findOneBy(['telegramId' => $this->bot->chatId()]);
        if (null === $chat) {
            $chat = new Chat();
            $chat->setTelegramId($this->bot->chatId());
            $chat->setBalance(30);
            $chat->setTitle($this->bot->chat()->title);

            em()->persist($chat);
            em()->flush();
        }

        return $chat;
    }

    private function question(): Question {
        return em()->getRepository(Question::class)->find($this->questionId);
    }

    public function start(Nutgram $bot) {
        em()->clear();

        if ($this->chat()->getBalance() <= 0) {
            $bot->sendMessage('У вас закончились монеты. Пополните баланс  командой /balance и ждем вас снова!');
            $this->end();

            return;
        }

        /** @var Question $question */
        $question = em()->getRepository(Question::class)->getQuestion($this->chat());
        if (null === $question) {
            $bot->sendMessage('Уоу, вы ответили на все вопросы! Приходите завтра! Новые интересные вопросы появляются каждый день!');
            if ($bot->chatId() !== (int)$_ENV['ADMIN_CHAT_ID']) {
                $bot->sendMessage(text: 'Достигнут конец вопросов в чате ' . $this->chat()->__toString(), chat_id: $_ENV['ADMIN_CHAT_ID']);
            }
            $this->end();

            return;
        }

        $this->questionId = $question->getId();

        if (null !== $question->getQuestionPicture()) {
            $picture = em()->getRepository(Picture::class)->find($question->getQuestionPicture()->getId());
            $photo = fopen($this->picture($picture->getPath()), 'rb');

            /** @var Message $message */
            $bot->sendPhoto(
                photo       : InputFile::make($photo),
                chat_id     : $bot->chatId(),
                caption     : $question->ask(),
                parse_mode  : ParseMode::MARKDOWN,
                reply_markup: $this->questionReplyMarkUp(),
            );
        } else {
            $bot->sendMessage(
                text        : $question->ask(),
                parse_mode  : ParseMode::MARKDOWN,
                reply_markup: $this->questionReplyMarkUp(),
            );
        }

        $this->next('waitResponse');
    }

    public function waitResponse(Nutgram $bot) {
        if (null === $bot->callbackQuery()) {
            $this->handleTextResponse($bot);

            return;
        }

        $this->action = $bot->callbackQuery()->data;

        if ($this->action === 'skip') {
            if (null === $reaction = $this->reaction()) {
                $reaction = new Reaction($this->chat(), $this->question());
                em()->persist($reaction);
            }
            $reaction->skip();
            em()->flush();
            $this->start($bot);
        } elseif ($this->action === 'fail') {
            $this->fail($bot);
        } elseif ($this->action === 'continue') {
            $this->start($bot);
        } elseif ($this->action === 'finish') {
            $this->finish($bot);
        }
    }

    public function handleTextResponse(Nutgram $bot) {
        if (mb_strtolower($bot->message()->text) === mb_strtolower($this->question()->getAnswer())) {
            if (null !== $this->question()->getAnswerPicture()) {
                $picture = em()->getRepository(Picture::class)->find($this->question()->getAnswerPicture()->getId());
                $photo = fopen($this->picture($picture->getPath()), 'rb');

                /** @var Message $message */
                $bot->sendPhoto(
                    photo       : InputFile::make($photo),
                    chat_id     : $bot->chatId(),
                    caption     : $this->response(),
                    parse_mode  : ParseMode::MARKDOWN,
                    reply_markup: $this->successResponseReplyMarkUp(),
                );
            } else {
                $bot->sendMessage(text        : $this->response(),
                                  parse_mode  : ParseMode::MARKDOWN,
                                  reply_markup: $this->successResponseReplyMarkUp());
            }

            if (null === $reaction = $this->reaction()) {
                $reaction = new Reaction($this->chat(), $this->question());
                em()->persist($reaction);
            }

            $this->chat()->decrease();
            $reaction->response();

            em()->flush();
        } else {
            $bot->sendMessage(
                text: 'Ответ неверный. Попробуйте еще раз'
            );
        }
    }

    public function reaction(): ?Reaction {
        return em()->getRepository(Reaction::class)
                   ->createQueryBuilder('r')
                   ->where('r.chat = :chat')
                   ->andWhere('r.question = :question')
                   ->setParameter('chat', $this->chat())
                   ->setParameter('question', $this->question())
                   ->getQuery()
                   ->getOneOrNullResult();
    }

    private function fail(Nutgram $bot) {
        if (null !== $this->question()->getAnswerPicture()) {
            $picture = em()->getRepository(Picture::class)->find($this->question()->getAnswerPicture()->getId());
            $photo = fopen($this->picture($picture->getPath()), 'rb');

            /** @var Message $message */
            $bot->sendPhoto(
                photo       : InputFile::make($photo),
                chat_id     : $bot->chatId(),
                caption     : $this->failMessage(),
                parse_mode  : ParseMode::MARKDOWN,
                reply_markup: $this->failReplyMarkUp(),
            );
        } else {
            $bot->sendMessage(text        : $this->failMessage(),
                              parse_mode  : ParseMode::MARKDOWN,
                              reply_markup: $this->failReplyMarkUp());
        }

        if (null === $reaction = $this->reaction($bot)) {
            $reaction = new Reaction($this->chat(), $this->question());
            em()->persist($reaction);
        }

        $this->chat()->decrease();
        $reaction->fail();
        em()->flush();

        $this->next('waitResponse');
    }

    private function finish(Nutgram $bot) {
        if (null !== $this->question()->getAnswerPicture()) {
            $picture = em()->getRepository(Picture::class)->find($this->question()->getAnswerPicture()->getId());
            $photo = fopen($this->picture($picture->getPath()), 'rb');

            /** @var Message $message */
            $bot->sendPhoto(
                photo     : InputFile::make($photo),
                chat_id   : $bot->chatId(),
                caption   : $this->failMessage(),
                parse_mode: ParseMode::MARKDOWN
            );
        } else {
            $bot->sendMessage(text      : $this->failMessage(),
                              parse_mode: ParseMode::MARKDOWN);
        }

        if (null === $reaction = $this->reaction($bot)) {
            $reaction = new Reaction($this->chat(), $this->question());
            em()->persist($reaction);
        }

        $this->chat()->decrease();
        $reaction->fail();
        em()->flush();

        $bot->sendMessage('Приходите завтра! Новые интересные вопросы появляются каждый день!');

        $this->end();
    }

    private function failReplyMarkUp(): InlineKeyboardMarkup {
        return InlineKeyboardMarkup::make()
                                   ->addRow(
                                       InlineKeyboardButton::make('Точно!', callback_data: 'continue'),
                                       InlineKeyboardButton::make('Ладно, хватит', callback_data: 'finish'),
                                   );
    }

    private function successResponseReplyMarkUp(): InlineKeyboardMarkup {
        return InlineKeyboardMarkup::make()
                                   ->addRow(
                                       InlineKeyboardButton::make('Продолжаем', callback_data: 'continue'),
                                       InlineKeyboardButton::make('Закончить', callback_data: 'finish'),
                                   );
    }

    private function failMessage(): string {
        $response = "*Правильный ответ:*\n" . $this->escape($this->question()->getAnswer());
        $response .= $this->comment();

        return $response;
    }

    private function response(): string {
        $response = "*Это правильный ответ\!*";
        $response .= $this->comment();

        return $response;
    }

    private function comment(): string {
        if (null !== $this->question()->getComment()) {
            return sprintf("\n\n%s", $this->escape($this->question()->getComment()));
        }

        return '';
    }

    private function questionReplyMarkUp(): InlineKeyboardMarkup {
        return InlineKeyboardMarkup::make()
                                   ->addRow(
                                       InlineKeyboardButton::make('Пропустить', callback_data: 'skip'),
                                       InlineKeyboardButton::make('Показать ответ', callback_data: 'fail'),
                                       InlineKeyboardButton::make('Закончить', callback_data: 'finish'),
                                   );
    }

    private function picture(string $path): string {
        return "https://storage.yandexcloud.net/qweasley/" . $path;
    }

    private function escape(string $text): string {
        foreach ([
                     '?',
                     '!',
                     '_',
                     '*',
                     '[',
                     ']',
                     '(',
                     ')',
                     '~',
                     '`',
                     '>',
                     '<',
                     '&',
                     '#',
                     '+',
                     '-',
                     '=',
                     '|',
                     '{',
                     '}',
                     '.'
                 ] as $symbol) {
            $text = str_replace($symbol, '\\' . $symbol, $text);
        }

        return $text;
    }
}
