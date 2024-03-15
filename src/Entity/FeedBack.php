<?php

namespace App\Entity;

use App\Repository\FeedBackRepository;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Table(name: 'feedbacks')]
#[ORM\Entity(repositoryClass: FeedBackRepository::class)]
#[ORM\HasLifecycleCallbacks]
class FeedBack {

    use CreatedAtTrait;

    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column(options: ['default' => 'nextval(\'feedbacks_id_seq\')'])]
    private ?int $id = null;

    #[ORM\Column(type: 'text', length: 4000)]
    private string $text;

    #[ORM\Column(type: 'text', length: 4000, nullable: true)]
    private ?string $response;

    #[ORM\ManyToOne(targetEntity: Chat::class, cascade: ['persist'], fetch: 'LAZY')]
    #[ORM\JoinColumn(nullable: false, onDelete: 'CASCADE')]
    private Chat $chat;

    public function getId(): ?int {
        return $this->id;
    }

    public function getChat(): Chat {
        return $this->chat;
    }

    public function setChat(Chat $chat): void {
        $this->chat = $chat;
    }

    public function getText(): string {
        return $this->text;
    }

    public function setText(string $text): void {
        $this->text = $text;
    }

    public function getResponse(): ?string {
        return $this->response;
    }

    public function setResponse(string $response): void {
        $this->response = $response;
    }
}
