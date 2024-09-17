CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE tender_status_type AS ENUM (
    'Created',
    'Published',
    'Closed'
);

CREATE TYPE service_type_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
);

CREATE TABLE tender (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status tender_status_type,
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_version INT
);

CREATE TABLE tender_version (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    service_type service_type_type,
    version INT,
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE
);

CREATE TYPE bid_status_type AS ENUM (    
    'Created',
    'Published',
    'Canceled',
    'Approved',
    'Rejected'
);

CREATE TYPE bid_decision_type AS ENUM (    
    'Approved',
    'Rejected',
    '-'
);

CREATE TYPE bid_author_type AS ENUM (
    'User',
    'Organization'
);

CREATE TABLE bid (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status bid_status_type,
    decision bid_decision_type default '-',
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
    author_id UUID REFERENCES employee(id),
    author_type bid_author_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_version INT
);

CREATE TABLE bid_version (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version INT,
    bid_id UUID REFERENCES bid(id) ON DELETE CASCADE
);

CREATE TABLE review (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    author_id UUID REFERENCES employee(id),
    receiver_id UUID REFERENCES employee(id),
    bid_id UUID REFERENCES bid(id) ON DELETE CASCADE
);

CREATE TABLE approves (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bid_id UUID REFERENCES bid(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employee(id) ON DELETE CASCADE
);
