// Run with: npm test
//
// Node's own test runner, strip-types and assert — no framework. The helpers are
// pure, so this needs neither a DOM nor a build step.
import assert from 'node:assert/strict'
import test from 'node:test'
import { matches, price, sellPrice } from './wiki-format.ts'

test('price picks the largest denomination that reads cleanly', () => {
  assert.equal(price(1), '1 cp')
  assert.equal(price(9), '9 cp')
  assert.equal(price(10), '1 sp')
  assert.equal(price(99), '9.9 sp')
  assert.equal(price(100), '1 gp')
  assert.equal(price(150), '1.5 gp')
  assert.equal(price(2000), '20 gp')
})

test('price does not eat the zeros of a round number', () => {
  // Trailing-zero trimming must not reach past the decimal point and turn
  // 1000 gp into "1".
  assert.equal(price(100000), '1000 gp')
  assert.equal(price(1000), '10 gp')
  assert.equal(price(2000000), '20000 gp')
})

test('price treats worthless and negative as free', () => {
  assert.equal(price(0), 'free')
  assert.equal(price(-5), 'free')
})

test('sellPrice is half, rounded down, floored at one copper', () => {
  assert.equal(sellPrice(100), 50)
  assert.equal(sellPrice(3), 1)
  assert.equal(sellPrice(1), 1, 'a 1 cp item still sells for something')
  assert.equal(sellPrice(0), 0, 'but a worthless item does not')
})

test('matches is case-insensitive across every field', () => {
  assert.equal(matches('gob', 'Goblin', 'humanoid'), true)
  assert.equal(matches('HUMAN', 'Goblin', 'humanoid'), true)
  assert.equal(matches('dragon', 'Goblin', 'humanoid'), false)
})

test('matches searches numbers too, so a CR or level is findable', () => {
  assert.equal(matches('12', 'Goblin', 12), true)
})

test('an empty or whitespace query matches everything', () => {
  assert.equal(matches('', 'Goblin'), true)
  assert.equal(matches('   ', 'Goblin'), true)
})
