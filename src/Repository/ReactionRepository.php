<?php

namespace App\Repository;

use App\Entity\Reaction;
use Doctrine\ORM\EntityRepository;

/**
 * @extends EntityRepository<Reaction>
 *
 * @method Reaction|null find($id, $lockMode = null, $lockVersion = null)
 * @method Reaction|null findOneBy(array $criteria, array $orderBy = null)
 * @method Reaction[]    findAll()
 * @method Reaction[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class ReactionRepository extends EntityRepository
{

    public function getReactedDql(): string {
        $reactionsQuery = $this->createQueryBuilder('r');
        $reactionsQuery->select('question.id')
                       ->innerJoin('r.question', 'question')
                       ->andWhere($reactionsQuery->expr()->eq('r.chat', ':chat'));

        return $reactionsQuery->getDQL();
    }

    public function getNotSkippedDql(): string {
        $reactionsQuery = $this->createQueryBuilder('r');
        $reactionsQuery->select('q.id')
                       ->innerJoin('r.question', 'q')
                       ->andWhere($reactionsQuery->expr()->eq('r.chat', ':chat'))
                       ->andWhere(
                           $reactionsQuery->expr()->orX(
                               $reactionsQuery->expr()->isNotNull('r.responsedAt'),
                               $reactionsQuery->expr()->isNotNull('r.failedAt')
                           ));

        return $reactionsQuery->getDQL();
    }
}
