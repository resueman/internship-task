CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE employee (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
);

CREATE TABLE organization (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type organization_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE
);

INSERT INTO employee (username, first_name, last_name) VALUES
    ('a1', 'a1 first name', 'a1 last name'),
    ('b1', 'b1 first name', 'b1 last name'),
    ('b2', 'b2 first name', 'b2 last name'),
    ('c1', 'c1 first name', 'c1 last name'),
    ('c2', 'c2 first name', 'c2 last name'),
    ('c3', 'c3 first name', 'c3 last name'),
    ('d1', 'd1 first name', 'd1 last name'),
    ('d2', 'd2 first name', 'd2 last name'),
    ('d3', 'd3 first name', 'd3 last name'),
    ('d4', 'd4 first name', 'd4 last name');

INSERT INTO organization (name, description, type) VALUES
    ('A organization', 'A description', 'IE'),
    ('B organization', 'B description', 'LLC'),
    ('C organization', 'C description', 'JSC'),
    ('D organization', 'D description', 'LLC');

INSERT INTO organization_responsible (user_id, organization_id)
SELECT e.id, o.id FROM ((SELECT id FROM employee WHERE username LIKE '%a%') e
CROSS JOIN
(SELECT id FROM organization WHERE name LIKE '%A%') o);

INSERT INTO organization_responsible (user_id, organization_id)
SELECT e.id, o.id FROM ((SELECT id FROM employee WHERE username LIKE '%b%') e
CROSS JOIN
(SELECT id FROM organization WHERE name LIKE '%B%') o);

INSERT INTO organization_responsible (user_id, organization_id)
SELECT e.id, o.id FROM ((SELECT id FROM employee WHERE username LIKE '%c%') e
CROSS JOIN
(SELECT id FROM organization WHERE name LIKE '%C%') o);

INSERT INTO organization_responsible (user_id, organization_id)
SELECT e.id, o.id FROM ((SELECT id FROM employee WHERE username LIKE '%d%') e
CROSS JOIN
(SELECT id FROM organization WHERE name LIKE '%D%') o);

---------------------------------------------------------------------------------------------

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
