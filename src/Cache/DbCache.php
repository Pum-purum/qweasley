<?php

namespace App\Cache;

use Psr\Cache\InvalidArgumentException;
use Psr\SimpleCache\CacheInterface;
use Symfony\Component\Cache\Adapter\PdoAdapter;

class DbCache implements CacheInterface {

    private PdoAdapter $cache;

    public function __construct(PdoAdapter $cache)
    {
        $this->cache = $cache;
    }


    public function get(string $key, mixed $default = null): mixed {
        return $this->cache->get($key, function() {
            return null;
        });
    }

    /**
     * @throws InvalidArgumentException
     */
    public function set(string $key, mixed $value, \DateInterval|int|null $ttl = null): bool {
        $item = $this->cache->getItem($key);
        $item->set($value);
        if ($ttl !== null) {
            $item->expiresAfter($ttl);
        }
        return $this->cache->save($item);
    }

    /**
     * @throws InvalidArgumentException
     */
    public function delete(string $key): bool {
        return $this->cache->deleteItem($key);
    }

    public function clear(): bool {
        return $this->cache->clear();
    }

    public function getMultiple(iterable $keys, mixed $default = null): iterable {
        return $this->cache->getItems((array)$keys);
    }

    /**
     * @throws InvalidArgumentException
     */
    public function setMultiple(iterable $values, \DateInterval|int|null $ttl = null): bool {
        foreach ($values as $key => $value) {
            $this->set($key, $value, $ttl);
        }

        return true;
    }

    /**
     * @throws InvalidArgumentException
     */
    public function deleteMultiple(iterable $keys): bool {
        return $this->cache->deleteItems((array)$keys);
    }

    /**
     * @throws InvalidArgumentException
     */
    public function has(string $key): bool {
        return $this->cache->hasItem($key);
    }
}
