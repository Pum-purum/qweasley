<?php

namespace App\Repository;

use App\Entity\Chat;
use App\Entity\Question;
use App\Entity\Reaction;
use Doctrine\ORM\EntityRepository;

/**
 * @extends EntityRepository<Question>
 *
 * @method Question|null find($id, $lockMode = null, $lockVersion = null)
 * @method Question|null findOneBy(array $criteria, array $orderBy = null)
 * @method Question[]    findAll()
 * @method Question[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class QuestionRepository extends EntityRepository {
    public function getQuestion(Chat $chat): ?Question {
        $reactionsRepository = em()->getRepository(Reaction::class);

        $qb = $this->createQueryBuilder('q');
        $qb->select('q.id')
            ->andWhere(
                $qb->expr()->orX(
                    $qb->expr()->neq('q.author', ':author'),
                    $qb->expr()->isNull('q.author')
                )
            )
            ->andWhere('q.isPublished = TRUE')
            ->andWhere($qb->expr()->notIn('q.id', $reactionsRepository->getReactedDql()));
        $qb->setParameter('chat', $chat);
        $qb->setParameter('author', $chat);
        $ids = $qb->getQuery()->getSingleColumnResult();

        // Если пусто, то повторяем те вопросы, что были пропущены
        if (empty($ids)) {
            $questionQueryBuilder = $this->createQueryBuilder('qq');
            $questionQueryBuilder->select('qq.id')
                ->andWhere(
                    $qb->expr()->orX(
                        $qb->expr()->neq('qq.author', ':author'),
                        $qb->expr()->isNull('qq.author')
                    )
                )
                ->andWhere('qq.isPublished = TRUE')
                ->andWhere($questionQueryBuilder->expr()->notIn('qq.id', $reactionsRepository->getNotSkippedDql()));
            $questionQueryBuilder->setParameter('chat', $chat);
            $questionQueryBuilder->setParameter('author', $chat);
            $ids = $questionQueryBuilder->getQuery()->getSingleColumnResult();
        }
        if (empty($ids)){
            return null;
        }

        shuffle($ids);

        return $this->find($ids[0]);
    }
}
