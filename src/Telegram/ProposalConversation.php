<?php

namespace App\Telegram;

use App\Entity\Chat;
use App\Entity\FeedBack;
use App\Entity\Picture;
use App\Entity\Question;
use App\Entity\Reaction;
use Aws\S3\S3Client;
use Doctrine\ORM\Exception\ORMException;
use Doctrine\ORM\OptimisticLockException;
use Psr\SimpleCache\InvalidArgumentException;
use SergiX44\Nutgram\Conversations\Conversation;
use SergiX44\Nutgram\Nutgram;
use SergiX44\Nutgram\Telegram\Types\Keyboard\InlineKeyboardButton;
use SergiX44\Nutgram\Telegram\Types\Keyboard\InlineKeyboardMarkup;
use SergiX44\Nutgram\Telegram\Types\Media\PhotoSize;

class ProposalConversation extends Conversation {

    public string $action = '';
    public bool $isPicture = false;
    public string $caption = '';
    public string $answer = '';
    public ?string $comment = null;
    public ?string $pictureQuestion = null;
    public ?string $pictureAnswer = null;

    public function start(Nutgram $bot) {
        $bot->sendMessage('Отличная идея! Каким будет ваш вопрос?',
            reply_markup: $this->startReplyMarkUp());

        $this->next('waitType');
    }

    public function waitType(Nutgram $bot) {
        if (null === $bot->callbackQuery()) {
            $bot->sendMessage('Что-то пошло не так...Давайте еще раз',
                reply_markup: $this->startReplyMarkUp());
            $this->start($bot);

            return;
        }

        $this->action = $bot->callbackQuery()->data;

        if ($this->action === 'text') {
            $this->bot->sendMessage('Введите текст вопроса');
            $this->next('textQuestion');
        } elseif ($this->action === 'picture') {
            $this->isPicture = true;
            $this->pictureQuestion();
        }
    }

    public function textQuestion() {
        $this->caption = $this->bot->message()->text;

        $this->bot->sendMessage('Введите ответ на вопрос');

        $this->next('waitAnswer');
    }

    public function waitAnswer(Nutgram $bot) {
        $this->answer = $this->bot->message()->text;

        $replyMarkUp = InlineKeyboardMarkup::make()
                                           ->addRow(
                                               InlineKeyboardButton::make('Закончить', callback_data: 'finish')
                                           );
        $message = 'Введите комментарий, который появится под ответом, или нажмите Закончить';
        if (true === $this->isPicture) {
            $replyMarkUp = InlineKeyboardMarkup::make()
                                              ->addRow(
                                                  InlineKeyboardButton::make('Пропустить', callback_data: 'suggestAnswerPicture')
                                              );
            $message = 'Введите текстовый комментарий, который появится под ответом, или нажмите Пропустить';
        }
        $this->bot->sendMessage($message, reply_markup: $replyMarkUp);

        $this->next('waitComment');
    }

    public function waitComment(Nutgram $bot) {
        if (null === $this->bot->callbackQuery()) {
            $this->comment = $this->bot->message()->text;

            if (true === $this->isPicture) {
                $this->suggestAnswerPicture();
            } else {
                $this->finish();
            }

            return;
        }

        $this->action = $this->bot->callbackQuery()->data;
        if ($this->action === 'finish') {
            $this->finish();
        } elseif ($this->action === 'suggestAnswerPicture') {
            $this->suggestAnswerPicture();
        }
    }

    public function suggestAnswerPicture(): void {
        $replyMarkUp = InlineKeyboardMarkup::make()
                                           ->addRow(
                                               InlineKeyboardButton::make('Закончить', callback_data: 'finish')
                                           );
        $message = 'Загрузите картинку-комментарий, которая появится вместе с ответом, или нажмите Закончить';
        $this->bot->sendMessage($message, reply_markup: $replyMarkUp);

        $this->next('waitAnswerPicture');
    }

    public function waitAnswerPicture(): void {
        if (null === $this->bot->callbackQuery()) {
            $this->pictureAnswer = $this->upload();
            $this->finish();

            return;
        }

        $this->action = $this->bot->callbackQuery()->data;
        if ($this->action === 'finish') {
            $this->finish();
        }
    }

    private function upload(): ?string {
        /** @var PhotoSize $photoSize */
        $photoSize = end($this->bot->message()->photo);

        if ($photoSize->file_size > 1024 * 1024) {
            $this->bot->sendMessage('Файл слишком большой. Попробуйте загрузить файл поменьше');
            $this->next('waitQuestionPicture');

            return null;
        }

        $s3 = new S3Client([
            'region'      => $_ENV['AWS_S3_REGION'],
            'version'     => 'latest',
            'credentials' => [
                'key'    => $_ENV['AWS_S3_ACCESS_KEY'],
                'secret' => $_ENV['AWS_S3_SECRET_KEY'],
            ],
            'endpoint'    => $_ENV['AWS_S3_ENTRYPOINT']
        ]);

        $file = $this->bot->getFile($photoSize->file_id);
        $exploded = explode('.', $file->url());
        $extension = end($exploded);
        $key = md5($file->url()) . '.' . $extension;
        $result = $s3->putObject([
            'Bucket'      => $_ENV['AWS_S3_BUCKET'],
            'Key'         => $key,
            'ContentType' => 'image/' . $extension,
            'Body'        => file_get_contents($file->url()),
        ]);

        if ($result['@metadata']['statusCode'] == 200) {
            return $key;
        } else {
            $this->bot->sendMessage('Что-то пошло не так. Повторите попытку позже');
            $this->end();

            return null;
        }
    }

    public function waitQuestionPicture(Nutgram $bot): void {
        if (!$this->bot->message()->photo) {
            $this->bot->sendMessage('Файл не распознан как картинка. Предложите вопрос снова');
            $this->end();

            return;
        }

        $this->pictureQuestion = $this->upload();

        $this->bot->sendMessage('Введите текст вопроса');
        $this->next('textQuestion');
    }

    /**
     * @throws OptimisticLockException
     * @throws ORMException
     * @throws InvalidArgumentException
     */
    public function finish(): void {
        $chat = em()->getRepository(Chat::class)->findOneBy(['telegramId' => $this->bot->chatId()]);

        $question = new Question($this->caption, $this->answer);
        $question->setAuthor($chat);

        if (null !== $this->comment) {
            $question->setComment($this->comment);
        }
        if (null !== $this->pictureQuestion) {
            $picture = new Picture();
            $picture->setPath($this->pictureQuestion);
            em()->persist($picture);
            $question->setQuestionPicture($picture);
        }
        if (null !== $this->pictureAnswer) {
            $picture = new Picture();
            $picture->setPath($this->pictureAnswer);
            em()->persist($picture);
            $question->setAnswerPicture($picture);
        }

        em()->persist($question);
        em()->flush();

        $message = 'Ваш вопрос принят. Спасибо! В случае одобрения вам придет уведомление в этот чат';
        $this->bot->sendMessage($message);

        $this->end();
    }

    public function pictureQuestion() {
        $message = 'Загрузите картинку-вопрос';
        $this->bot->sendMessage($message);

        $this->next('waitQuestionPicture');
    }

    private function startReplyMarkUp(): InlineKeyboardMarkup {
        return InlineKeyboardMarkup::make()
                                   ->addRow(
                                       InlineKeyboardButton::make('Текстовый', callback_data: 'text'),
                                       InlineKeyboardButton::make('С картинкой', callback_data: 'picture')
                                   );
    }
}
