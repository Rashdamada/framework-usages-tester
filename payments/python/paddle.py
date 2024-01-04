from paddle import PaddleClient

# https://github.com/paddle-python/paddle-client?tab=readme-ov-file#usage
paddle = PaddleClient(vendor_id=12345, api_key='myapikey')

# Checkout API
paddle.get_order_details(checkout_id='aaaa-bbbb-cccc-1234')
paddle.get_user_history(email='test@example.com')
paddle.get_prices(product_ids=[1234])

# Product API
paddle.list_coupons(product_id=1234)
paddle.create_coupon(
    coupon_type='product',
    discount_type='percentage',
    discount_amount=50,
    allowed_uses=1,
    recurring=False,
    currency='USD',
    product_ids=[1234],
    coupon_code='50%OFF',
    description='50% off coupon over $10',
    expires='2030-01-01 10:00:00',
    minimum_threshold=10,
    group='paddle-python',
)
paddle.delete_coupon(coupon_code='mycoupon', product_id=1234)
paddle.update_coupon(
    coupon_code='mycoupon',
    new_coupon_code='40%OFF',
    new_group='paddle-python-test',
    product_ids=[1234],
    expires='2030-01-01 10:00:00',
    allowed_uses=1,
    currency='USD',
    minimum_threshold=10,
    discount_amount=40,
    recurring=True
)
paddle.list_products()
paddle.list_transactions(entity='subscription', entity_id=1234)
paddle.refund_product_payment(order_id=1234, amount=0.01, reason='reason')

# Subscription API
paddle.list_plans()
paddle.get_plan(plan=123)
paddle.create_plan(
    plan_name='plan_name',
    plan_trial_days=14,
    plan_length=1,
    plan_type='month',
    main_currency_code='USD',
    initial_price_usd=50,
    recurring_price_usd=50,
)
paddle.list_subscription_users()
paddle.cancel_subscription(subscription_id=1234)
paddle.update_subscription(subscription_id=1234, pause=True)
paddle.update_subscription(
    subscription_id=1234,
    quantity=10.00,
    currency='USD',
    recurring_price=10.00,
    bill_immediately=False,
    plan_id=123,
    prorate=True,
    keep_modifiers=True,
    passthrough='passthrough',
)
paddle.pause_subscription(subscription_id=1234)
paddle.resume_subscription(subscription_id=1234)
paddle.add_modifier(subscription_id=1234, modifier_amount=10.5)
paddle.delete_modifier(modifier_id=10)
paddle.list_modifiers()
paddle.list_subscription_payments()
paddle.reschedule_subscription_payment(payment_id=4567, date='2030-01-01')
paddle.create_one_off_charge(
    subscription_id=1234,
    amount=0.0,
    charge_name="Add X on top of subscription"
)

# Alert API
paddle.get_webhook_history()
