<?php

use App\Cache\DbCache;
use App\Telegram\BalanceConversation;
use App\Telegram\FeedBackConversation;
use App\Telegram\ProposalConversation;
use App\Telegram\RulesConversation;
use App\Telegram\StartConversation;
use Doctrine\DBAL\DriverManager;
use Doctrine\ORM\EntityManager;
use Doctrine\ORM\Mapping\UnderscoreNamingStrategy;
use Doctrine\ORM\ORMSetup;
use SergiX44\Nutgram\Configuration;
use SergiX44\Nutgram\Nutgram;
use Symfony\Component\Cache\Adapter\PdoAdapter;

global $entityManager, $password;

require_once "vendor/autoload.php";

function createEm(): EntityManager {
    $paths = ['./'];

    // the connection configuration
    $dbParams = [
        'driver'      => 'pgsql',
        'driverClass' => 'Doctrine\DBAL\Driver\PDO\PgSQL\Driver',
        'host'        => $_ENV['DB_HOST'],
        'user'        => $_ENV['DB_USER'],
        'port'        => $_ENV['DB_PORT'],
        'password'    => $_ENV['DB_PASS'],
        'dbname'      => $_ENV['DB_NAME'],
        'sslmode'     => 'require',
    ];
    $namingStrategy = new UnderscoreNamingStrategy(CASE_UPPER);
    $config = ORMSetup::createAttributeMetadataConfiguration($paths, false);
    $config->setNamingStrategy($namingStrategy);

    $connection = DriverManager::getConnection($dbParams, $config);

    return new EntityManager($connection, $config);
}

function em(): EntityManager {
    global $entityManager, $password;
    if (!$entityManager || !$entityManager->isOpen()) {
        $entityManager = createEm();
    }

    return $entityManager;
}

function handler($payload, $context) {
    global $password;
    $password = $context->getToken()->getAccessToken();
    $params = em()->getConnection()->getParams();
    $cache = new PdoAdapter(
        sprintf('%s:host=%s;port=%s;dbname=%s;sslmode=require', $params['driver'], $params['host'], $params['port'], $params['dbname']),
        '',
        3600,
        [
            'db_username' => $params['user'],
            'db_password' => $params['password'],
        ]
    );
    $dbCache = new DbCache($cache);
    $bot = new Nutgram($_ENV['TELEGRAM_TOKEN'], new Configuration(
        cache: $dbCache
    ));

    $bot->onCommand('start', StartConversation::class)->description('Начать квиз');
    $bot->onCommand('balance', BalanceConversation::class)->description('Баланс');
    $bot->onCommand('rules', RulesConversation::class)->description('Правила');
    $bot->onCommand('feedback', FeedBackConversation::class)->description('Обратная связь');
    $bot->onCommand('proposal', ProposalConversation::class)->description('Предложить вопрос');

    $bot->onCommand('start@qweasleybot', StartConversation::class)->description('Начать квиз');
    $bot->onCommand('balance@qweasleybot', BalanceConversation::class)->description('Баланс');
    $bot->onCommand('rules@qweasleybot', RulesConversation::class)->description('Правила');
    $bot->onCommand('feedback@qweasleybot', FeedBackConversation::class)->description('Обратная связь');
    $bot->onCommand('proposal@qweasleybot', ProposalConversation::class)->description('Предложить вопрос');

    $bot->run();
}
