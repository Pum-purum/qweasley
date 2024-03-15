<?php

namespace App\Entity;

use App\Repository\ReactionRepository;
use Doctrine\DBAL\Types\Types;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Table(name: 'reactions')]
#[ORM\Entity(repositoryClass: ReactionRepository::class)]
#[ORM\UniqueConstraint(name: 'chat_question', columns: ['chat_id', 'question_id'])]
#[ORM\HasLifecycleCallbacks]
class Reaction {

    use CreatedAtTrait;

    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column(name: 'id', type: Types::INTEGER, options: ['default' => 'nextval(\'reactions_id_seq\')'])]
    private ?int $id = null;

    #[ORM\Column(type: 'datetime_immutable', nullable: true)]
    private ?\DateTimeImmutable $responsedAt = null;

    #[ORM\Column(type: 'datetime_immutable', nullable: true)]
    private ?\DateTimeImmutable $skippedAt = null;

    #[ORM\Column(type: 'datetime_immutable', nullable: true)]
    private ?\DateTimeImmutable $failedAt = null;

    #[ORM\ManyToOne(targetEntity: Chat::class, cascade: ['persist'], fetch: 'LAZY')]
    #[ORM\JoinColumn(referencedColumnName: 'id', nullable: false, onDelete: 'CASCADE')]
    private Chat $chat;

    #[ORM\ManyToOne(targetEntity: Question::class, cascade: ['persist'], fetch: 'LAZY')]
    #[ORM\JoinColumn(referencedColumnName: 'id', nullable: false, onDelete: 'CASCADE')]
    private Question $question;

    public function __construct(Chat $chat, Question $question) {
        $this->chat = $chat;
        $this->question = $question;
    }

    public function getId(): ?int {
        return $this->id;
    }

    public function responsedAt(): ?\DateTimeImmutable {
        return $this->responsedAt;
    }

    public function skippedAt(): ?\DateTimeImmutable {
        return $this->skippedAt;
    }

    public function failedAt(): ?\DateTimeImmutable {
        return $this->failedAt;
    }

    public function response(): self {
        $this->responsedAt = new \DateTimeImmutable();

        return $this;
    }

    public function skip(): self {
        $this->skippedAt = new \DateTimeImmutable();

        return $this;
    }

    public function fail(): self {
        $this->failedAt = new \DateTimeImmutable();

        return $this;
    }

    public function getChat(): Chat {
        return $this->chat;
    }

    public function getQuestion(): Question {
        return $this->question;
    }
}
