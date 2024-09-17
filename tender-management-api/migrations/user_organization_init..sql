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