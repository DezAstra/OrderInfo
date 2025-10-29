// static/script.js

document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('orderForm');
    const orderUidInput = document.getElementById('orderUid');
    const orderDataSection = document.getElementById('orderDataSection');
    const orderDataDiv = document.getElementById('orderData');
    const errorMessageSection = document.getElementById('errorMessageSection');
    const errorMessageDiv = document.getElementById('errorMessage');

    form.addEventListener('submit', async (event) => {
        event.preventDefault(); // Предотвращаем стандартную отправку формы

        const orderUid = orderUidInput.value.trim();
        if (!orderUid) {
            showError('Пожалуйста, введите UID заказа.');
            return;
        }

        // Очищаем предыдущие данные и ошибки
        hideOrderData();
        hideError();

        try {
            // Делаем AJAX-запрос к HTTP-эндпоинту сервиса
            // Предполагаем, что сервис слушает на том же хосте и порту, что и веб-интерфейс
            const response = await fetch(`/order/${encodeURIComponent(orderUid)}`);
            
            if (!response.ok) {
                if (response.status === 404) {
                    throw new Error('Заказ с таким UID не найден.');
                } else {
                    throw new Error(`Ошибка HTTP: ${response.status} ${response.statusText}`);
                }
            }

            const order = await response.json();

            // Отображаем данные заказа
            displayOrder(order);
        } catch (error) {
            console.error('Ошибка при получении заказа:', error);
            showError(error.message || 'Произошла ошибка при получении заказа.');
        }
    });

    function displayOrder(order) {
        let html = `
            <div class="order-details">
                <h3>Основная информация</h3>
                <dl>
                    <dt>UID заказа:</dt><dd>${escapeHtml(order.order_uid)}</dd>
                    <dt>Трек-номер:</dt><dd>${escapeHtml(order.track_number)}</dd>
                    <dt>Вход:</dt><dd>${escapeHtml(order.entry)}</dd>
                    <dt>Язык:</dt><dd>${escapeHtml(order.locale)}</dd>
                    <dt>Внутренняя подпись:</dt><dd>${escapeHtml(order.internal_signature || '')}</dd>
                    <dt>ID клиента:</dt><dd>${escapeHtml(order.customer_id)}</dd>
                    <dt>Сервис доставки:</dt><dd>${escapeHtml(order.delivery_service)}</dd>
                    <dt>Ключ шарда:</dt><dd>${escapeHtml(order.shardkey)}</dd>
                    <dt>ID SM:</dt><dd>${escapeHtml(order.sm_id.toString())}</dd>
                    <dt>Дата создания:</dt><dd>${escapeHtml(order.date_created)}</dd>
                    <dt>OOF шард:</dt><dd>${escapeHtml(order.oof_shard)}</dd>
                </dl>
            </div>
        `;

        // Данные доставки
        html += `
            <div class="order-details">
                <h3>Информация о доставке</h3>
                <dl>
                    <dt>Имя:</dt><dd>${escapeHtml(order.delivery.name)}</dd>
                    <dt>Телефон:</dt><dd>${escapeHtml(order.delivery.phone)}</dd>
                    <dt>Индекс:</dt><dd>${escapeHtml(order.delivery.zip)}</dd>
                    <dt>Город:</dt><dd>${escapeHtml(order.delivery.city)}</dd>
                    <dt>Адрес:</dt><dd>${escapeHtml(order.delivery.address)}</dd>
                    <dt>Регион:</dt><dd>${escapeHtml(order.delivery.region)}</dd>
                    <dt>Email:</dt><dd>${escapeHtml(order.delivery.email)}</dd>
                </dl>
            </div>
        `;

        // Данные платежа
        html += `
            <div class="order-details">
                <h3>Информация о платеже</h3>
                <dl>
                    <dt>Транзакция:</dt><dd>${escapeHtml(order.payment.transaction)}</dd>
                    <dt>ID запроса:</dt><dd>${escapeHtml(order.payment.request_id || '')}</dd>
                    <dt>Валюта:</dt><dd>${escapeHtml(order.payment.currency)}</dd>
                    <dt>Провайдер:</dt><dd>${escapeHtml(order.payment.provider)}</dd>
                    <dt>Сумма:</dt><dd>${escapeHtml(order.payment.amount.toString())}</dd>
                    <dt>Дата платежа (Unix):</dt><dd>${escapeHtml(order.payment.payment_dt.toString())}</dd>
                    <dt>Банк:</dt><dd>${escapeHtml(order.payment.bank)}</dd>
                    <dt>Стоимость доставки:</dt><dd>${escapeHtml(order.payment.delivery_cost.toString())}</dd>
                    <dt>Общая стоимость товаров:</dt><dd>${escapeHtml(order.payment.goods_total.toString())}</dd>
                    <dt>Пользовательская комиссия:</dt><dd>${escapeHtml(order.payment.custom_fee.toString())}</dd>
                </dl>
            </div>
        `;

        // Товары
        if (order.items && order.items.length > 0) {
            html += `
                <div class="order-details">
                    <h3>Товары</h3>
                    <div class="table-wrapper"> <!-- Для горизонтальной прокрутки на мобильных -->
                        <table class="items-table">
                            <thead>
                                <tr>
                                    <th>ID товара (chrt_id)</th>
                                    <th>Трек-номер</th>
                                    <th>Цена</th>
                                    <th>RID</th>
                                    <th>Название</th>
                                    <th>Скидка (%)</th>
                                    <th>Размер</th>
                                    <th>Итоговая цена</th>
                                    <th>NM ID</th>
                                    <th>Бренд</th>
                                    <th>Статус</th>
                                </tr>
                            </thead>
                            <tbody>
            `;
            order.items.forEach(item => {
                html += `
                    <tr>
                        <td>${escapeHtml(item.chrt_id.toString())}</td>
                        <td>${escapeHtml(item.track_number)}</td>
                        <td>${escapeHtml(item.price.toString())}</td>
                        <td>${escapeHtml(item.rid)}</td>
                        <td>${escapeHtml(item.name)}</td>
                        <td>${escapeHtml(item.sale.toString())}</td>
                        <td>${escapeHtml(item.size)}</td>
                        <td>${escapeHtml(item.total_price.toString())}</td>
                        <td>${escapeHtml(item.nm_id.toString())}</td>
                        <td>${escapeHtml(item.brand)}</td>
                        <td>${escapeHtml(item.status.toString())}</td>
                    </tr>
                `;
            });
            html += `
                            </tbody>
                        </table>
                    </div>
                </div>
            `;
        } else {
            html += `<div class="order-details"><p>Товары не найдены.</p></div>`;
        }

        orderDataDiv.innerHTML = html;
        showOrderData();
    }

    function showError(message) {
        errorMessageDiv.textContent = message;
        errorMessageSection.classList.remove('hidden');
    }

    function hideError() {
        errorMessageSection.classList.add('hidden');
        errorMessageDiv.textContent = '';
    }

    function showOrderData() {
        orderDataSection.classList.remove('hidden');
    }

    function hideOrderData() {
        orderDataSection.classList.add('hidden');
        orderDataDiv.innerHTML = '';
    }

    // Функция для экранирования HTML, чтобы предотвратить XSS
    function escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '<',
            '>': '>',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.replace(/[&<>"']/g, m => map[m]);
    }
});
