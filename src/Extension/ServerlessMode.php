<?php

namespace App\Extension;

use SergiX44\Nutgram\RunningMode\Webhook;

class ServerlessMode extends Webhook {

    protected function input(): ?string
    {
        global $updates;

        return $updates ?: null;
    }
}
