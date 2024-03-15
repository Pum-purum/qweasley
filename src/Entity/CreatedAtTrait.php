<?php

namespace App\Entity;

use Doctrine\ORM\Mapping as ORM;

trait CreatedAtTrait {

    #[ORM\Column(type: 'datetime_immutable', nullable: false)]
    private ?\DateTimeImmutable $createdAt = null;

    public function createdAt(): \DateTimeImmutable {
        return $this->createdAt;
    }

    #[ORM\PrePersist]
    public function prePersist(): void {
        $this->createdAt = new \DateTimeImmutable();
    }
}
