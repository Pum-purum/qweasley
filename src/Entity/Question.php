<?php

namespace App\Entity;

use App\Repository\QuestionRepository;
use Doctrine\DBAL\Types\Types;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Table(name: 'questions')]
#[ORM\Entity(repositoryClass: QuestionRepository::class)]
#[ORM\HasLifecycleCallbacks]
class Question {

    use CreatedAtTrait;

    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column(name: 'id', type: Types::INTEGER, options: ['default' => 'nextval(\'questions_id_seq\')'])]
    private ?int $id = null;

    #[ORM\Column(type: 'text', length: 4000)]
    private string $text;

    #[ORM\Column(type: 'text', length: 4000)]
    private string $answer;

    #[ORM\Column(type: 'text', length: 4000, nullable: true)]
    private ?string $comment;

    #[ORM\ManyToOne(targetEntity: Chat::class, fetch: 'LAZY')]
    #[ORM\JoinColumn(nullable: true)]
    private ?Chat $author = null;

    #[ORM\Column(type: 'boolean', options: ['default' => false])]
    private bool $isPublished = false;

    #[ORM\OneToOne(targetEntity: Picture::class, cascade: ['persist', 'remove'], fetch: 'LAZY', orphanRemoval: true)]
    #[ORM\JoinColumn(nullable: true, onDelete: 'SET NULL')]
    private ?Picture $questionPicture = null;

    #[ORM\OneToOne(targetEntity: Picture::class, cascade: ['persist', 'remove'], fetch: 'LAZY', orphanRemoval: true)]
    #[ORM\JoinColumn(nullable: true)]
    private ?Picture $answerPicture = null;

    #[ORM\Column(type: 'datetime_immutable', nullable: true)]
    private ?\DateTimeImmutable $approvedAt = null;

    public function __construct(string $text, string $answer) {
        $this->text = $text;
        $this->answer = $answer;
        $this->isPublished = false;
    }

    public function getId(): ?int {
        return $this->id;
    }

    public function getText(): string {
        return $this->text;
    }
    public function getAnswer(): string {
        return $this->answer;
    }

    public function setText(string $text): Question {
        $this->text = $text;

        return $this;
    }

    public function setAnswer(string $answer): Question {
        $this->answer = $answer;

        return $this;
    }

    public function isPublished(): bool {
        return $this->isPublished;
    }

    public function setIsPublished(bool $isPublished): Question {
        if ($this->isPublished && $this->approvedAt === null) {
            $this->approvedAt = new \DateTimeImmutable();
        }

        $this->isPublished = $isPublished;

        return $this;
    }

    public function getQuestionPicture(): ?Picture {
        return $this->questionPicture;
    }

    public function getAuthor(): ?Chat {
        return $this->author;
    }

    public function setAuthor(?Chat $author): void {
        $this->author = $author;
    }

    public function getComment(): ?string {
        return $this->comment;
    }

    public function setComment(?string $comment): void {
        $this->comment = $comment;
    }

    public function getAnswerPicture(): ?Picture {
        return $this->answerPicture;
    }

    public function setAnswerPicture(?Picture $answerPicture): void {
        $this->answerPicture = $answerPicture;
    }

    public function setQuestionPicture(?Picture $questionPicture): void {
        $this->questionPicture = $questionPicture;
    }

    public function getApprovedAt(): ?\DateTimeImmutable {
        return $this->approvedAt;
    }

    public function setApprovedAt(?\DateTimeImmutable $approvedAt): void {
        $this->approvedAt = $approvedAt;
    }
}
