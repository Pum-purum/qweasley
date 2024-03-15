<?php

namespace App\Repository;

use App\Entity\FeedBack;
use Doctrine\ORM\EntityRepository;

/**
 * @extends EntityRepository<FeedBack>
 *
 * @method FeedBack|null find($id, $lockMode = null, $lockVersion = null)
 * @method FeedBack|null findOneBy(array $criteria, array $orderBy = null)
 * @method FeedBack[]    findAll()
 * @method FeedBack[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class FeedBackRepository extends EntityRepository
{
}
