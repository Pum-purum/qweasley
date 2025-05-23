<?php

use App\Cache\DbCache;
use App\Extension\ServerlessMode;
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

global $entityManager;

require_once "vendor/autoload.php";

function createEm(): EntityManager {

    $paths = [__DIR__];

    // the connection configuration
    $dbParams = [
        'driver'      => 'postgres',
        'driverClass' => 'Doctrine\DBAL\Driver\PDO\PgSQL\Driver',
        'host'        => $_ENV['DB_HOST'],
        'user'        => $_ENV['DB_USER'],
        'port'        => $_ENV['DB_PORT'],
        'password'    => $_ENV['DB_PASSWORD'],
        'dbname'      => $_ENV['DB_NAME'],
        'sslmode'     => 'require',
        'sslrootcert' => '/etc/ssl/certs/ca-certificates.crt'
    ];
    $namingStrategy = new UnderscoreNamingStrategy(CASE_UPPER);
    $config = ORMSetup::createAttributeMetadataConfiguration($paths, true);
    $config->setNamingStrategy($namingStrategy);

    $connection = DriverManager::getConnection($dbParams, $config);

    return new EntityManager($connection, $config);
}

function em(): EntityManager {
    global $entityManager;

    if (!$entityManager || !$entityManager->isOpen()) {
        $entityManager = createEm();
    }

    return $entityManager;
}

function handler($payload, $context) {
    global $updates;

    try {
        $updates = $payload['body'];
        $params = em()->getConnection()->getParams();

        $cache = new PdoAdapter(
            sprintf('pgsql:host=%s;port=%s;dbname=%s;sslmode=%s;sslrootcert=%s', $params['host'], $params['port'], $params['dbname'], $params['sslmode'], $params['sslrootcert']),
            '',
            0,
            [
                'db_username' => $params['user'],
                'db_password' => $params['password'],
                'db_connection_options' => [
                    PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
                    PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
                    PDO::ATTR_EMULATE_PREPARES   => true,
                ]
            ]
        );

        $dbCache = new DbCache($cache);
        $bot = new Nutgram($_ENV['TELEGRAM_TOKEN'], new Configuration(
            cache: $dbCache
        ));
        $bot->setRunningMode(ServerlessMode::class);
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
    } catch (Exception $e) {
        return [
            'statusCode' => 400,
            'body'       => $e->getMessage()
        ];
    }

    return [
        'statusCode' => 200,
        'body'       => ''
    ];
}
