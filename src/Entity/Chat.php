<?php

namespace App\Entity;

use App\Repository\ChatRepository;
use Doctrine\DBAL\Types\Types;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Table(name: 'chats')]
#[ORM\Entity(repositoryClass: ChatRepository::class)]
#[ORM\HasLifecycleCallbacks]
class Chat {

    use CreatedAtTrait;

    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'AUTO')]
    #[ORM\Column(name: 'id', type: Types::INTEGER, unique: true, nullable: false, options: ['default' => 'nextval(\'chats_id_seq\')'])]
    private ?int $id = null;

    #[ORM\Column(type: 'integer')]
    private int $balance = 0;

    #[ORM\Column(type: Types::BIGINT, unique: true, nullable: false)]
    private int $telegramId;

    #[ORM\Column(type: 'text', length: 255, nullable: true)]
    private ?string $title;

    public function __construct() {

    }

    public function getId(): ?int {

        return $this->id;
    }

    public function getBalance(): int {
        return $this->balance;
    }

    public function setBalance(int $balance): Chat {
        $this->balance = $balance;

        return $this;
    }

    public function decrease(): Chat {
        $this->balance--;

        return $this;
    }

    public function __toString(): string {
        return ' (' . ($this->id ?? '-') . ')' . ($this->title ?? '-');
    }

    public function getTitle(): ?string {
        return $this->title;
    }

    public function getTelegramId(): int {
        return $this->telegramId;
    }

    public function setTitle(?string $title): Chat {
        $this->title = $title;
        return $this;
    }

    public function setTelegramId(int $telegramId): void {
        $this->telegramId = $telegramId;
    }
}
