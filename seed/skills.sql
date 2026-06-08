-- Canonical skill registry seed data
-- Aliases are used for normalisation: "golang" → "go"

INSERT INTO skills (id, name, category, aliases) VALUES
  -- Languages
  (uuid_generate_v4(), 'go', 'language', ARRAY['golang', 'go lang', 'go programming']),
  (uuid_generate_v4(), 'python', 'language', ARRAY['python3', 'python2', 'py']),
  (uuid_generate_v4(), 'java', 'language', ARRAY['java 11', 'java 17', 'java 21', 'core java']),
  (uuid_generate_v4(), 'javascript', 'language', ARRAY['js', 'node.js', 'nodejs', 'node']),
  (uuid_generate_v4(), 'typescript', 'language', ARRAY['ts']),
  (uuid_generate_v4(), 'rust', 'language', ARRAY['rust-lang']),
  (uuid_generate_v4(), 'cpp', 'language', ARRAY['c++', 'c plus plus']),
  (uuid_generate_v4(), 'scala', 'language', ARRAY[]),

  -- DSA
  (uuid_generate_v4(), 'dsa', 'dsa', ARRAY['data structures', 'algorithms', 'data structures and algorithms', 'competitive programming', 'leetcode']),
  (uuid_generate_v4(), 'dynamic-programming', 'dsa', ARRAY['dp', 'dynamic programming']),
  (uuid_generate_v4(), 'graphs', 'dsa', ARRAY['graph algorithms', 'bfs', 'dfs', 'graph theory']),
  (uuid_generate_v4(), 'trees', 'dsa', ARRAY['binary trees', 'bst', 'binary search tree', 'tree traversal']),

  -- Backend
  (uuid_generate_v4(), 'kafka', 'backend', ARRAY['apache kafka', 'kafka consumer groups', 'kafka streams']),
  (uuid_generate_v4(), 'redis', 'backend', ARRAY['redis cache', 'redis pub/sub', 'redis cluster']),
  (uuid_generate_v4(), 'postgresql', 'backend', ARRAY['postgres', 'psql', 'pg']),
  (uuid_generate_v4(), 'mysql', 'backend', ARRAY['mariadb']),
  (uuid_generate_v4(), 'rest-api', 'backend', ARRAY['rest', 'restful api', 'http api']),
  (uuid_generate_v4(), 'grpc', 'backend', ARRAY['grpc', 'protobuf', 'protocol buffers']),
  (uuid_generate_v4(), 'docker', 'devops', ARRAY['containerization', 'containers']),
  (uuid_generate_v4(), 'kubernetes', 'devops', ARRAY['k8s', 'kube']),
  (uuid_generate_v4(), 'aws', 'devops', ARRAY['amazon web services', 'aws cloud']),
  (uuid_generate_v4(), 'git', 'devops', ARRAY['github', 'gitlab', 'version control']),

  -- System Design
  (uuid_generate_v4(), 'system-design', 'system_design', ARRAY['system design', 'distributed systems design', 'hld', 'high level design']),
  (uuid_generate_v4(), 'microservices', 'system_design', ARRAY['microservice architecture', 'soa', 'service oriented']),
  (uuid_generate_v4(), 'distributed-systems', 'system_design', ARRAY['distributed computing', 'distributed databases']),
  (uuid_generate_v4(), 'caching', 'system_design', ARRAY['cache design', 'caching strategies', 'cdn']),
  (uuid_generate_v4(), 'message-queues', 'system_design', ARRAY['message queue', 'event streaming', 'pub-sub', 'rabbitmq']),
  (uuid_generate_v4(), 'load-balancing', 'system_design', ARRAY['load balancer', 'reverse proxy', 'nginx']),

  -- Architecture
  (uuid_generate_v4(), 'architecture', 'architecture', ARRAY['software architecture', 'system architecture', 'technical architecture']),
  (uuid_generate_v4(), 'lld', 'architecture', ARRAY['low level design', 'object oriented design', 'ood', 'design patterns']),
  (uuid_generate_v4(), 'clean-architecture', 'architecture', ARRAY['hexagonal architecture', 'ddd', 'domain driven design']),

  -- Domain
  (uuid_generate_v4(), 'payments', 'domain', ARRAY['payment systems', 'payment gateway', 'upi', 'fintech']),
  (uuid_generate_v4(), 'ecommerce', 'domain', ARRAY['e-commerce', 'marketplace', 'cart', 'checkout']),
  (uuid_generate_v4(), 'machine-learning', 'domain', ARRAY['ml', 'deep learning', 'ai', 'neural networks'])

ON CONFLICT (name) DO NOTHING;
