local config = import 'default.jsonnet';

config {
  'omini_9002-1'+: {
    'app-config'+: {
      pruning: 'everything',
      'state-sync'+: {
        'snapshot-interval': 0,
      },
    },
    genesis+: {
      app_state+: {
        feemarket+: {
          params+: {
            min_gas_multiplier: '0',
          },
        },
      },
    },
  },
}
